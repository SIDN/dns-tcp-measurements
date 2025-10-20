# TODO for querier
TODO: 
- Make sure we can set the source port and IP address (5) 
- Is it better to have one client shared by all goroutines or many clients? (7)
- Probably something else is also needed, but not sure what. (?8)

DONE:
- Make a resolve function that takes as input one dns.Msg and queries the auth NS for this 
		query. It then takes the response and checks whether it is truncated. If it is it needs 
		to resend the message. (1)
- Think about connections, do I need to keep TCP connections open? (5) -> Answer: no, just 
        don't think about it
- Make a query sender that goes over a list of queries and calls 
		upon the resolve function to send this query in parallel i guess (2)
- Make a function that reads the query input file and turns it into a slice of query messages
- Extend the query sender such that it sends queries with the right timing (if necessary). (5)


NOTE:
- Normaal stond ulimit -n op 1024, om het goed te laten werken heb ik eerst dit verhoogd naar 4096, maar dat was nog niet genoeg dus nu is het verhoogd naar 8192 en daarna zelfs naar 16384. Hierdoor kreeg ik niet meer workers, maar ik kreeg wel dat alles een stuk sneller ging
- Interessant: als je een tijdje het niet gerund hebt, dan lijkt het dat er iets op slaapstand gaat en dus wat langer doet over reageren. Want dan gaat hij opeens naar 5000 per seconde. Dit neemt dan af en wordt steeds sneller als je de query vaker stuurt tot hij weer op rond de 20000 qps uitkomt. -> dit is gefixt door pasta ipv slirp4netns te gebruiken.