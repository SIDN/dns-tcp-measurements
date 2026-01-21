package main

import (
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"math/rand"
	"os"
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

// A Query holds the necessary information to make a DNS message
// relative offset that we want this query to have in the replay
type Query struct {
	queryStrings []string      // strings used to construct actual DNS query
	offset       time.Duration // offset that is to be used during replay
}

// createDNSMsg returns a *dns.Msg created based on the query info stored
// in the `queryStrings` []string. If an error occurs during this process, it returns the
// corresponding `error`.
func createDNSMsg(queryStrings []string) (*dns.Msg, error) {
	// protocol := record[3] //TODO check what do we do with the protocol: UDP/TCP (all should be UDP right?)
	request := queryStrings[0]
	// Handle special ".4o." (aka .nl.) query by changing it to a "4o." query.
	request = strings.TrimPrefix(request, ".")

	reqType := queryStrings[1]

	// The dns library is not able to handle the A6 and TYPE97 types
	// so we replace them
	if reqType == "A6" {
		request = "64hpx3g.4o."
		reqType = "A"
	} else if reqType == "TYPE97" {
		request = "64hpx3g.4o."
		reqType = "A"
	}

	DOBit := queryStrings[3]
	m := new(dns.Msg)

	if value, exists := dns.StringToType[reqType]; exists {
		dnsutil.SetQuestion(m, dnsutil.Fqdn(request), value)
		if DOBit == "1" {
			m.Security = true
		}
	} else {
		return m, fmt.Errorf("createDNSMsg: The qtype %s does not exist", reqType)
	}

	return m, nil
}

// readCsvFile returns a slice of string slices that contains the records
// and corresponding fields of the csv file stored at `filePath`.
func readCsvFile(filePath string) ([][]string, error) {
	// Code from https://stackoverflow.com/questions/24999079/reading-csv-file-in-go
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("readCsvFile: Unable to read input file for %s with error %s", filePath, err)
	}
	defer f.Close()

	csvReader := csv.NewReader(f)
	records, err := csvReader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("readCsvFile: Unable to parse file as CSV for %s with error %s", filePath, err)
	}

	return records, nil
}

// readQueryData returns a slice filled with Query's that the csv
// file stored at `filename` contains. Each line of the csv should
// look like  offset,protocol,domainname,requesttype. If an error
// occurs, it returns the corresponding `error`.
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

		timing, err := time.ParseDuration(offsetStr)
		if err != nil {
			return nil, fmt.Errorf("readQueryData: error while parsing the offset for offset: %s", offsetStr)
		}
		// To save memory the Query does not contain an actual *dns.Msg, but only the necessary
		// information to create one. The actual message is created just before sending it.
		query := Query{queryStrings: record[1:], offset: timing}
		queries = append(queries, query)
	}

	return queries, nil
}

// resolve returns a Response that it gets from the nameserver at
// `address` when it queries for DNS question `m` using `client`.
func resolve(m *dns.Msg, address string, client *dns.Client, percentage float64) Response {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	resp, rtt, err := client.Exchange(ctx, m, "udp", address)
	tcp := false

	if err != nil {
		return Response{err: fmt.Errorf("resolve: error while doing udp exchange: %s", err)}
	}
	if rand.Float64()*100 < percentage { //According to rand we make sure we only select `percentage` of queries
		tcp = true
		resp, rtt, err = client.Exchange(ctx, m, "tcp", address)
		if err != nil {
			return Response{err: fmt.Errorf("resolve: error while doing tcp exchange: %s", err)}
		}
	}

	return Response{resp: resp, rtt: rtt, err: nil, tcp: tcp}
}

// SendQueries is a function that has a number of goroutines take a query from the
// `queries` channel, send it with the right timing to `address`. The response it gets
// it will put in the `responses` channel. The parameter `percentage` defines what
// percentage of the DNS queries needs to be retried over TCP.
func SendQueries(queries <-chan Query, address string, responses chan<- Response, percentage float64) {
	defer close(responses)
	sem := make(chan struct{}, 165) // A channel to enable concurrency with 165 workers
	client := new(dns.Client)       // Reuse one DNS client
	start := time.Now()
	var wg sync.WaitGroup        //WaitGroup to make sure we wait until everything is finished
	for query := range queries { //run until there are no more queries in the queries channel
		sem <- struct{}{} // get a worker slot
		wg.Add(1)
		go func(q Query) {
			defer wg.Done()
			defer func() { <-sem }()                 // release worker slot
			sleep := time.Until(start.Add(q.offset)) // Only send message once we have gotten to the right time
			if sleep > 0 {
				time.Sleep(sleep)
			}
			msg, err := createDNSMsg(q.queryStrings)
			if err != nil {
				fmt.Printf("%s\n", fmt.Errorf("SendQueries: error while creating dns message: %s", err))
			}
			response := resolve(msg, address, client, percentage)

			if response.err != nil {
				fmt.Printf("SendQueries: error while resolving: %s\n", response.err.Error())
			}
			responses <- response
		}(query)
	}

	wg.Wait() //wait until all workers are done
	close(sem)
}

func main() {
	queryFilename := flag.String("f", "test-csv/test_file_structure.csv", "Filename specifying what file the queries should be read from.")
	percentage := flag.Float64("p", 100.0, "A float value that represents the percentage of queries that we want to retry over TCP.")
	nameserverAddress := flag.String("s", "127.0.0.1:4242", "A string in the format of [IP-address]:[port] that specifies at what address the nameserver is listening.")
	flag.Parse() //Get the command line arguments

	queries, err := readQueryData(*queryFilename)
	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}

	// Now we want to turn the Query slice into a Query channel
	queryCh := make(chan Query, len(queries))
	for _, q := range queries {
		queryCh <- q
	}
	numQueries := len(queries)
	close(queryCh) //No more messages will come after

	responseCh := make(chan Response, 1000) // Limit the memory used by not trying to save all responses, but handle them concurrently
	fmt.Println("Done with prep")
	var duration time.Duration

	// Start the SendQueries async to make sure we can handle the responses sent to responseCh concurrently
	go func() {
		start := time.Now()
		SendQueries(queryCh, *nameserverAddress, responseCh, *percentage)
		duration = time.Since(start)
	}()
	rcodeCounter := make(map[uint16]int)

	counter := 0
	actualQueries := 0 //This is the number of queries that is actually sent, accounting for the double query in tcp
	tcp := 0
	for response := range responseCh { // This loop then handles the responses concurrently
		if response.err == nil {
			counter++
			rcodeCounter[response.resp.Rcode]++
			actualQueries++
			if response.tcp {
				tcp++
				actualQueries++
			}
		}
	}
	fmt.Printf("Succesful responses: %d\n", counter)
	fmt.Printf("Total queries: %d\n", numQueries)
	fmt.Printf("Amount with TCP: %d \n", tcp)
	fmt.Printf("Total queries, incl. duplicate TCP: %d\n", actualQueries)
	fmt.Printf("Execution time: %s\n", duration)
	fmt.Println("\nEncountered Rcodes and their count")
	for rcode, count := range rcodeCounter {
		fmt.Printf("RCode %s, Count %d\n", dns.RcodeToString[rcode], count)
	}
}
