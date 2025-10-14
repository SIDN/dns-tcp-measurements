# DNS querier
Querier for replaying DNS queries against an authoritative nameserver. \
It should be possible to do both a fast replay (send the DNS queries as fast as possible) and a 
timing replay (send the queries with the relative timing provided in the input file)

## Input
The dataformat of the input file should be as follows:
```
time,qname,qtype,prot
0.001,example.nl,AAAA,UDP
```
Where `offset` is the time in seconds relative to the start of the replay at which the query should be sent. The other headers 
are pretty straightforward, so if it should use UDP or TCP, what domain name it should request and what type of 
record it should request. 

## Build and usage
Build with `go install gitlab.sidnlabs.nl/eline.stehouwer/querier`, then run with `~/go/bin/querier`. \
Before running the code it is important that you also have a nameserver running that you run the code against, 
if you want to test it against a local nameserver. 

### Building local nameserver
In this repository, under `nsd`, we have put a simple Dockerfile that installs an `nsd` nameserver and the corresponding 
file structure needed to set this nameserver up. To get this nameserver up and running you only need to provide 
the zone files yourself (and edit the .conf file to point to these zone files) and build the container as follows:
```
podman build -f Dockerfile --tag=querier-nsd:latest
```
Then afterwards you can run you the `nsd` nameserver in a container with the following command (provided you have the 
right file structure)
```
podman run --rm -it -v ./nsd.conf:/etc/nsd/nsd.conf -v ./zones:/dns -v ./config:/config -p 4242:53/tcp -p 4242:53/udp --name nsd-query querier-nsd:latest nsd -V 2 –d 
```
After you have this running in a container you can run the `querier` as explained above. 

## Needed configuration on machine
On the machine where you run this code, you need to make sure you can open up enough file descriptors. This can done by 
running:
```
ulimit -n 16384
```

## Note on performance
Right now it seems that the first time you run `~/go/bin/querier` the performance is relatively bad, if you then run it 
twice or thrice more it seems to speed up by up to 4 times. So this makes it seems there is some type of warmup needed. 
Keep this in mind when running the code.  
