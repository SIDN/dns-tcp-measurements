package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"math/rand"

	// "log"
	"os"
	// "sort"
	"flag"
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
	tcp  bool
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
	a6Dropped := 0
	t97Droped := 0
	for i, record := range records {
		if i == 1000000 {
			break //TODO comment once we do want to use all the queries
		}
		offsetStr := record[0]

		// Check whether the offsetStr ends on an `s`, if not we want to add it
		if !strings.HasSuffix(offsetStr, "s") {
			offsetStr = offsetStr + "ms" //We are working with milliseconds
		}

		// protocol := record[3] //TODO check what do we do with the protocol: UDP/TCP (all should be UDP right?)
		request := record[1]
		// Change requests to .nl. to a request to nl.
		request = strings.TrimPrefix(request, ".")

		reqType := record[2]
		if reqType == "A6" {
			a6Dropped++
			request = "example.nl."
			reqType = "NS"
			continue
		} else if reqType == "TYPE97" {
			t97Droped++
			request = "example.nl."
			reqType = "NS"
			continue
		}
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
	fmt.Printf("In total replaced %d queries, these consisted of: \n", a6Dropped+t97Droped)
	fmt.Printf("- %d A6 queries \n", a6Dropped)
	fmt.Printf("- %d TYPE97 queries \n", t97Droped)

	return queries, nil
}

// resolve returns a Response that it gets from the nameserver at
// address when it queries for DNS question m using client.
func resolve(m *dns.Msg, address string, client *dns.Client, percentage float64) Response {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	resp, rtt, err := client.Exchange(ctx, m, "udp", address)
	tcp := false
	// Redo:
	if err != nil {
		return Response{err: fmt.Errorf("resolve: error while doing udp exchange: %s", err)}
	}
	if rand.Float64()*100 < percentage { //According to rand we make sure we only select `percentage` of queries
		// fmt.Println("Got reponse with TC=1, so retrying over TCP")
		tcp = true
		resp, rtt, err = client.Exchange(ctx, m, "tcp", address)
		// goto Redo
		if err != nil {
			return Response{err: fmt.Errorf("resolve: error while doing tcp exchange: %s", err)}
		}
	}

	return Response{resp: resp, rtt: rtt, err: nil, tcp: tcp}
}

// SendQueries is a function that has a number of goroutines take a query from the
// queries channel, send it with the right timing to address. The response it gets
// it will put in the responses channel
func SendQueries(queries <-chan Query, address string, responses chan<- Response, percentage float64) {
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
			response := resolve(q.quer, address, client, percentage)
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

	queryFilename := flag.String("f", "test-csv/test_file_structure.csv", "Filename specifying what file the queries should be read from.")
	percentage := flag.Float64("p", 100.0, "A float value that represents the percentage of queries that we want to retry over TCP.")
	nameserverAddress := flag.String("s", "127.0.0.1:4242", "A string in the format of [IP-address]:[port] that specifies at what address the nameserver is listening.")
	flag.Parse() //Get the command line arguments

	queries, err := readQueryData(*queryFilename)
	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}
	// sort.Sort(ByOffset(queries)) //Only necessary if the .csv file comes in unsorted

	// Now we want to turn the Query slice into a Query channel
	queryCh := make(chan Query, len(queries))
	for _, q := range queries {
		queryCh <- q
	}
	numQueries := len(queries)
	close(queryCh) //No more messages will come after
	responseCh := make(chan Response, len(queries))
	fmt.Println("Done with prep")
	start := time.Now()
	SendQueries(queryCh, *nameserverAddress, responseCh, *percentage)
	duration := time.Since(start)
	rcodeCounter := make(map[uint16]int)

	counter := 0
	actualQueries := 0 //This is the number of queries that is actually sent, accounting for the double query in tcp
	tcp := 0
	for response := range responseCh {
		if response.err == nil {
			counter++
			rcodeCounter[response.resp.Rcode]++
			actualQueries++
			if response.tcp {
				tcp++
				actualQueries++
			}
			//fmt.Printf("query time: %.3d µs, size: %d bytes\n", response.rtt/1e3, response.resp.Len())
		}
	}
	fmt.Printf("Number of error-less responses: %d of %d queries \n", counter, numQueries)
	fmt.Printf("Number of responses over TCP: %d \n", tcp)
	fmt.Printf("Number of queries, accounting for double query of TCP: %d\n", actualQueries)
	fmt.Printf("Execution time: %s\n", duration)
	fmt.Println("\nEncountered Rcodes and their count:")
	for rcode, count := range rcodeCounter {
		fmt.Printf("RCode: %s, Count: %d\n", dns.RcodeToString[rcode], count)
	}
}
