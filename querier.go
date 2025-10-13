package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"codeberg.org/miekg/dns"
)

type Response struct {
	resp *dns.Msg
	rtt  time.Duration
	err  error
}

type Query struct {
	quer   *dns.Msg
	offset time.Duration
}

// ByOffset implements the sort.Interface for []Query based on the
// offset field
type ByOffset []Query

func (o ByOffset) Len() int           { return len(o) }
func (o ByOffset) Swap(i, j int)      { o[i], o[j] = o[j], o[i] }
func (o ByOffset) Less(i, j int) bool { return o[i].offset < o[j].offset }

func createDNSMsg(domain string, qtype string, qclass uint16) (*dns.Msg, error) {
	m := new(dns.Msg)
	m.ID = dns.ID() //Make sure we have a random Query ID

	// Parse the question type
	switch strings.ToUpper(qtype) {
	case "A":
		m.Question = []dns.RR{&dns.A{Hdr: dns.Header{Name: domain, Class: qclass}}}
	case "AAAA":
		m.Question = []dns.RR{&dns.AAAA{Hdr: dns.Header{Name: domain, Class: qclass}}}
	case "MX":
		m.Question = []dns.RR{&dns.MX{Hdr: dns.Header{Name: domain, Class: qclass}}}
	case "CNAME":
		m.Question = []dns.RR{&dns.CNAME{Hdr: dns.Header{Name: domain, Class: qclass}}}
	case "TXT":
		m.Question = []dns.RR{&dns.TXT{Hdr: dns.Header{Name: domain, Class: qclass}}}
	default:
		return nil, fmt.Errorf("unsupported question type: %s", qtype)
	}

	return m, nil
}

func createQueryWithOffset(dnsMsg *dns.Msg, offset string) (Query, error) {
	d, err := time.ParseDuration(offset)
	if err != nil {
		fmt.Printf("bad offset %q: %v", offset, err)
		return Query{}, err
	}
	return Query{quer: dnsMsg, offset: d}, nil
}

func readCsvFile(filePath string) [][]string {
	// Code from https://stackoverflow.com/questions/24999079/reading-csv-file-in-go
	f, err := os.Open(filePath)
	if err != nil {
		log.Fatal("Unable to read input file "+filePath, err)
	}
	defer f.Close()

	csvReader := csv.NewReader(f)
	records, err := csvReader.ReadAll()
	if err != nil {
		log.Fatal("Unable to parse file as CSV for "+filePath, err)
	}

	return records
}

func readQueryData(filename string) ([]Query, error) {
	queries := []Query{}

	records := readCsvFile(filename)

	for _, record := range records {
		offsetStr := record[0]

		// Check whether the offsetStr ends on an `s`, if not we want to add it
		if !strings.HasSuffix(offsetStr, "s") {
			offsetStr = offsetStr + string('s')
		}

		// protocol := record[1] //TODO check what do we do with the protocol: UDP/TCP (all should be UDP right?)
		request := record[2] //TODO check and if necessary make it fqdn

		// Check whether request ends on a . if not add it to make it FQDN
		if !strings.HasSuffix(request, ".") {
			request = request + string('.')
		}

		reqType := record[3]
		msg, err := createDNSMsg(request, reqType, dns.ClassINET)
		if err != nil {
			fmt.Printf("Error while creating DNS Msg for request: %s, with error: %s\n", request, err)
			return queries, err
		}
		query, err := createQueryWithOffset(msg, offsetStr)
		if err != nil {
			fmt.Printf("Error while creating query with offset for request %s, with offset %s, and error %s\n", request, offsetStr, err)
			return queries, err
		}
		queries = append(queries, query)
	}

	return queries, nil
}

// m *Msg is a DNS question
// address is a string in the form of ip_addr:port which contains the IP address of the auth NS to query
func resolve(m *dns.Msg, address string, client *dns.Client) Response {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	resp, rtt, err := client.Exchange(ctx, m, "udp", address)

Redo:
	switch err {
	case nil:
		//Do nothing here
	default:
		fmt.Printf("%s\n", err.Error())
		return Response{resp: resp, rtt: rtt, err: err}
	}
	if resp.Truncated { //If the response is truncated then we want to start a TCP connection
		// fmt.Println("Got reponse with TC=1, so retrying over TCP")
		resp, rtt, err = client.Exchange(ctx, m, "tcp", address)
		goto Redo
	}

	return Response{resp: resp, rtt: rtt, err: err}
}

func SendQueries(queries <-chan Query, address string, responses chan<- Response) {
	defer close(responses)
	sem := make(chan struct{}, 165) //We cannot seem to make more workers without problems
	client := new(dns.Client)       // Reuse one client
	start := time.Now()
	var wg sync.WaitGroup
	for query := range queries { //run until there are no more queries in the queries channel
		sem <- struct{}{} // get a worker slot
		wg.Add(1)
		go func(q Query) {
			defer wg.Done()
			defer func() { <-sem }() // release worker slot
			sleep := time.Until(start.Add(q.offset))
			if sleep > 0 {
				time.Sleep(sleep)
			}

			fmt.Printf("Query sent at: %s (relative to start)\n", time.Since(start))
			response := resolve(q.quer, address, client)
			fmt.Printf("Query resolved at: %s (relative to start)\n", time.Since(start))
			if response.err != nil {
				fmt.Printf("%s\n", response.err.Error())
			}
			responses <- response
			// <- sem
		}(query)
	}

	wg.Wait() //wait until all workers are done
	close(sem)
}

func main() {
	/********** START GENERAL TEST FOR SIMPLE resolve BEHAVIOR **********/
	// testMsg, err := createDNSQuery("dhx5plx.de.", "A", dns.ClassINET)
	// if err != nil {
	// 	log.Print(err)
	// 	return
	// }
	// response := resolve(testMsg, "127.0.0.1:4241")
	// fmt.Println(response.rtt) //Takes 1.37 ms
	// testMsgLong, err := createDNSQuery("dhx5plx.de.", "TXT", dns.ClassINET)
	// response = resolve(testMsgLong, "127.0.0.1:4241")
	// fmt.Println(response.rtt) //Takes 1.95 ms
	/********** END GENERAL TEST FOR SIMPLE resolve BEHAVIOR **********/

	/********** START CODE TO CREATE QUERY WITH OFFSET **********
	testMsg, err := createDNSMsg("dhx5plx.de.", "A", dns.ClassINET)
	if err != nil {
		log.Print(err)
		return
	}
	offset := "0.4s"
	q, err := createQueryWithOffset(testMsg, offset)
	if err != nil {
		log.Print(err)
		return
	}
	********** STOP CODE TO CREATE QUERY WITH OFFSET **********/

	queries, err := readQueryData("test-csv/queries.csv")
	if err != nil {
		fmt.Printf("Error in readQueryData: %s\n", err)
		return
	}
	sort.Sort(ByOffset(queries))

	// Now we want to turn the Query slice into a Query channel
	queryCh := make(chan Query, len(queries))
	for _, q := range queries {
		queryCh <- q
	}

	close(queryCh) //No more messages will come after
	responseCh := make(chan Response, len(queries))
	start := time.Now()
	SendQueries(queryCh, "127.0.0.1:4242", responseCh)
	duration := time.Since(start)
	counter := 0
	for response := range responseCh {
		// fmt.Println("RESPONSE:")
		// fmt.Printf("%v", response.resp)
		if response.err != nil {
			fmt.Printf("%s\n", response.err.Error())
		} else {
			counter++
			//fmt.Printf("query time: %.3d µs, size: %d bytes\n", response.rtt/1e3, response.resp.Len())
		}
	}
	fmt.Printf("Number of responses: %d \n", counter)
	fmt.Printf("Execution time: %s\n", duration)
}
