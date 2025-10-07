# TODO for querier
TODO: 
- Make a function that reads the query input file and turns it into a slice of query messages
		for this it might be nice to make sure the input file is formated in the same format as 
		as what Msg.String retuns, because then we can use https://codeberg.org/miekg/dns/src/branch/main/dnsutil/msg.go 
		StringToMsg function to turn the string in the file into a DNS query. (3)
- Extend the resolve function to be able to handle the timing that is added to each of our 
		query entries in the query input file (4)
- Make a query sender that goes over a list of queries and calls 
		upon the resolve function to send this query in parallel i guess (2)
- Extend the query sender such that it sends queries with the right timing (if necessary). (5)
- Make sure we can set the source port and IP address (6) 
- Is it better to have one client shared by all goroutines or many clients? (8)
- Probably something else is also needed, but not sure what. (?9)

DONE:
- Make a resolve function that takes as input one dns.Msg and queries the auth NS for this 
		query. It then takes the response and checks whether it is truncated. If it is it needs 
		to resend the message. (1)
- Think about connections, do I need to keep TCP connections open? (7) -> Answer: no, just 
        don't think about it

