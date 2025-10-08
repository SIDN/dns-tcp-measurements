package main

import (
	"fmt"
	"context"
	"codeberg.org/miekg/dns"
	"time"
	"strings"
	"log"
	"sync"
)


type Response struct {
	resp *dns.Msg
	rtt time.Duration
	err error 
}

func createDNSQuery(domain string, qtype string, qclass uint16) (*dns.Msg, error) {
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


//m *Msg is a DNS question
//address is a string in the form of ip_addr:port which contains the IP address of the auth NS to query 
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

func sendQueries(queries <- chan *dns.Msg, address string, responses chan <- Response) {
	defer close(responses)
	sem := make(chan struct{}, 165) //We cannot seem to make more workers without problems
	client := new(dns.Client) // Reuse one client
	var wg sync.WaitGroup
	for query := range queries { //run until there are no more queries in the queries channel
		sem <- struct{}{} // get a worker slot
		wg.Add(1)
		go func(q *dns.Msg) {
			defer wg.Done()
			response := resolve(q, address, client)
			if response.err != nil {
				fmt.Printf("%s\n", response.err.Error())
			}
			responses <- response
			<- sem// release worker slot
		}(query)
	}

	wg.Wait() //wait until all workers are done
	close(sem)
}

func main(){
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

	testMsg, err := createDNSQuery("dhx5plx.de.", "A", dns.ClassINET)
	if err != nil {
		log.Print(err)
		return
	}
	lengthQueryCh := 20000 //The number of queries we want to send
	queryCh := 	make(chan *dns.Msg, lengthQueryCh)
	for i:= 0; i < lengthQueryCh; i++ {
		queryCh <- testMsg
	}
	close(queryCh) //No more messages will come after
	responseCh := make(chan Response, lengthQueryCh)
	start := time.Now()
	sendQueries(queryCh, "127.0.0.1:4242", responseCh)
	duration := time.Since(start)
	counter := 0
	for response := range responseCh {
		// fmt.Println("RESPONSE:")
		// fmt.Printf("%v", response.resp)
		if response.err != nil{
			fmt.Printf("%s\n", response.err.Error())
		} else {
			counter++
			//fmt.Printf("query time: %.3d µs, size: %d bytes\n", response.rtt/1e3, response.resp.Len())
		}
	}
	fmt.Printf("Number of responses: %d \n", counter)
    fmt.Printf("Execution time: %s\n", duration)
}
