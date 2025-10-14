package main

import (
	"context"
	"encoding/csv"
	"fmt"

	// "log"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"codeberg.org/miekg/dns"
	"codeberg.org/miekg/dns/dnsutil"
)

// A Response is the representation of a DNS response and corresponding
// metadata
type Response struct {
	resp *dns.Msg      // actual DNS response
	rtt  time.Duration // round trip time to the nameserver
	err  error         // any error that occurred
}

// A Query is the representation of a DNS query and the corresponding
// relative offset that we want this query to have in the replay
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

// createDNSMsg returns a *dns.Msg created based on domainname domain and
// query type qtype. If an error occurs during this process, it returns the
// corresponding error.
func createDNSMsg(domain string, qtype string) (*dns.Msg, error) {
	m := new(dns.Msg)

	if value, exists := dns.StringToType[qtype]; exists {
		dnsutil.SetQuestion(m, dnsutil.Fqdn(domain), value)
	} else {
		return m, fmt.Errorf("createDNSMsg: The qtype %s does not exist", qtype)
	}

	return m, nil
}

// createQueryWithOffset returns a Query, that consist of the given dnsMsg
// and the given offset. If an error occurs, it returns the corresponding
// error.
func createQueryWithOffset(dnsMsg *dns.Msg, offset string) (Query, error) {
	d, err := time.ParseDuration(offset)
	if err != nil {
		return Query{}, fmt.Errorf("createQueryWithOffset: bad offset %q: %v", offset, err)
	}
	return Query{quer: dnsMsg, offset: d}, nil
}

// readCsvFile returns a slice of string slices that contains the records
// and corresponding fields of the csv file stored at filePath.
func readCsvFile(filePath string) ([][]string, error) {
	// Code from https://stackoverflow.com/questions/24999079/reading-csv-file-in-go
	f, err := os.Open(filePath)
	if err != nil {
		// log.Fatal("Unable to read input file "+filePath, err)
		return nil, fmt.Errorf("readCsvFile: Unable to read input file for %s with error %s", filePath, err)
	}
	defer f.Close()

	csvReader := csv.NewReader(f)
	records, err := csvReader.ReadAll()
	if err != nil {
		// log.Fatal("Unable to parse file as CSV for "+filePath, err)
		return nil, fmt.Errorf("readCsvFile: Unable to parse file as CSV for %s with error %s", filePath, err)
	}

	return records, nil
}

// readQueryData returns a slices filled with Query's that the csv
// file stored at filename contains. Each line of the csv should
// look like  offset,protocol,domainname,requesttype. If an error
// occurs, it returns the corresponding error.
func readQueryData(filename string) ([]Query, error) {
	queries := []Query{}

	records, err := readCsvFile(filename)
	if err != nil {
		return nil, fmt.Errorf("readQueryData: %s", err)
	}

	for _, record := range records {
		offsetStr := record[0]

		// Check whether the offsetStr ends on an `s`, if not we want to add it
		if !strings.HasSuffix(offsetStr, "s") {
			offsetStr = offsetStr + "ms" //We are working with milliseconds
		}

		// protocol := record[3] //TODO check what do we do with the protocol: UDP/TCP (all should be UDP right?)
		request := record[1]
		reqType := record[2]
		msg, err := createDNSMsg(request, reqType)
		if err != nil {
			return nil, fmt.Errorf("readQueryData: error while creating DNS Msg for request: %s, with error: %s", request, err)
		}
		query, err := createQueryWithOffset(msg, offsetStr)
		if err != nil {
			return nil, fmt.Errorf("readQueryData: error while creating query with offset for request %s, with offset %s, and error %s", request, offsetStr, err)
		}
		queries = append(queries, query)
	}

	return queries, nil
}

// resolve returns a Response that it gets from the nameserver at
// address when it queries for DNS question m using client.
func resolve(m *dns.Msg, address string, client *dns.Client) Response {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	resp, rtt, err := client.Exchange(ctx, m, "udp", address)

Redo:
	if err != nil {
		return Response{err: fmt.Errorf("resolve: error while doing exchange: %s", err)}
	}
	if resp.Truncated { //If the response is truncated then we want to start a TCP connection
		// fmt.Println("Got reponse with TC=1, so retrying over TCP")
		resp, rtt, err = client.Exchange(ctx, m, "tcp", address)
		goto Redo
	}

	return Response{resp: resp, rtt: rtt, err: nil}
}

// SendQueries is a function that has a number of goroutines take a query from the
// queries channel, send it with the right timing to address. The response it gets
// it will put in the responses channel
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

			// fmt.Printf("Query sent at: %s (relative to start)\n", time.Since(start))
			response := resolve(q.quer, address, client)
			// fmt.Printf("Query resolved at: %s (relative to start)\n", time.Since(start))
			if response.err != nil {
				fmt.Printf("SendQueries: error while resolving: %s\n", response.err.Error())
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
	var query_filename string
	if len(os.Args) <= 1 {
		query_filename = "test-csv/test_file_structure.csv"
	} else {
		query_filename = os.Args[1]
	}
	queries, err := readQueryData(query_filename)
	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}
	sort.Sort(ByOffset(queries)) //Only necessary if the .csv file comes in unsorted

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
	rcodeCounter := make(map[uint16]int)

	counter := 0
	for response := range responseCh {
		if response.err == nil {
			counter++
			rcodeCounter[response.resp.Rcode]++
			//fmt.Printf("query time: %.3d µs, size: %d bytes\n", response.rtt/1e3, response.resp.Len())
		}
	}
	fmt.Printf("Number of error-less responses: %d \n", counter)
	fmt.Printf("Execution time: %s\n", duration)
	fmt.Println("\nEncountered Rcodes and their count:")
	for rcode, count := range rcodeCounter {
		fmt.Printf("RCode: %s, Count: %d\n", dns.RcodeToString[rcode], count)
	}
}
