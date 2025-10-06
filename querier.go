package main

import (
	"fmt"
	"context"
	"codeberg.org/miekg/dns"
	"time"
	"strings"
	"log"
)

// TODO: 
// - Make a function that reads the query input file and turns it into a slice of query messages
// 		for this it might be nice to make sure the input file is formated in the same format as 
//		as what Msg.String retuns, because then we can use https://codeberg.org/miekg/dns/src/branch/main/dnsutil/msg.go 
//		StringToMsg function to turn the string in the file into a DNS query. (3)
// - Make a resolve function that takes as input one dns.Msg and queries the auth NS for this 
//		query. It then takes the response and checks whether it is truncated. If it is it needs 
// 		to resend the message. (1)
// - Extend the resolve function to be able to handle the timing that is added to each of our 
//		query entries in the query input file (4)
// - Make a query sender that goes over all the queries that we read from the input file and calls 
// 		upon the resolve function to send this query with the right timing (if necessary). (2)
// - Probably something else is also needed, but not sure what. (?5)


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
//address is a string in the form of ip_addr:port which contains the IP address
//of the auth NS to query 
func resolve(m *dns.Msg, address string) (r *dns.Msg, rtt time.Duration, err error){ //TODO check if rtt is really relevant to return here
	//TODO: check if it is smarter to create the client inside this function or create it outside 
	//and pass it, so maybe we can reuse it
	c := new(dns.Client)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second) //TODO maybe create the background context in the main function?
	defer cancel() //TODO check whether this is correct, it is done as documented in https://pkg.go.dev/context#WithTimeout
	resp, rtt, err := c.Exchange(ctx, m, "udp", address)
	usedNet := "udp"
Redo:
	switch err {
	case nil:
		//Do nothing here
	default:
		fmt.Printf(";; %s\n", err.Error()) //TODO check if this is nice output format
		return resp, rtt, err 
	} 
	if resp.Truncated { //If the response is truncated then we want to start a TCP connection
		fmt.Println("Got reponse with TC=1, so retrying over TCP")
		resp, rtt, err = c.Exchange(ctx, m, "tcp", address)
		usedNet = "tcp"
		goto Redo
	}
	if resp.ID != m.ID {
		fmt.Println("Id mismatch\n")
		return resp, rtt, fmt.Errorf("Response Id %d does not match Message Id %d", resp.ID, m.ID)
	}


	fmt.Printf("%v", resp)
	fmt.Printf("\n;; query time: %.3d µs, server: %s(%s), size: %d bytes\n", rtt/1e3, address, usedNet, resp.Len())
	return resp, rtt, err
}

func main(){
	testMsg, err := createDNSQuery("dhx5plx.de.", "A", dns.ClassINET)
	if err != nil {
		log.Print(err)
		return
	}
	_, rtt, err := resolve(testMsg, "127.0.0.1:4241") 
	fmt.Println(rtt)
	testMsgLong, err := createDNSQuery("dhx5plx.de.", "TXT", dns.ClassINET)
	_, rtt, err = resolve(testMsgLong, "127.0.0.1:4241") 
	fmt.Println(rtt)
}
