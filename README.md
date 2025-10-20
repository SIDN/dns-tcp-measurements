# DNS querier
Querier for replaying DNS queries against an authoritative nameserver. The replay of these queries is done by trying to recreate the timing of each message.

## Input
The dataformat of the input file should be as follows:
```
0.001,example.nl,AAAA,UDP
```
This is in the following order: `time,qname,qtype,prot`. Where `offset` is the time in seconds relative to the start of the replay at which the query should be sent. The other types 
are pretty straightforward, so if it should use UDP or TCP, what domain name it should request and what type of 
record it should request. 

## Build and usage
Build with `go install gitlab.sidnlabs.nl/eline.stehouwer/querier`, then run with `~/go/bin/querier`. \
Before running the code it is important that you also have a nameserver running that you run the code against, 
if you want to test it against a local nameserver.  

There are several arguments that you can pass to the executable. These are:
- `-f [filename]`: with the `-f` option you can pass the input `.csv` file that holds the queries you want to send. 
The default value is: `test-csv/test_file_structure.csv`.
- `-p [percentage]`: with the `-p` option you can pass a number that represents the percentage of times that you 
want to retry over TCP. The default value is `100`.
- `-s [serveraddress]`: with the `-s` option you can pass the address of the name server that you want to send the queries 
to. The address is to be formatted in the following way: `[IP address]:[port]`. The default value is `127.0.0.1:4242`.

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
podman run --rm --network=pasta -it -v ./nsd.conf:/etc/nsd/nsd.conf -v ./zones:/dns -v ./config:/config -p 4242:53/tcp -p 4242:53/udp --name nsd-query querier-nsd:latest nsd -V 2 -d 
```
In this command we specify that the network used must be `pasta`. This is because the `slirp4netns` network seems to be slower and is less able to handle many TCP requests compared to the `pasta` network. \
After you have this running in a container you can run the `querier` as explained above. 

## Needed configuration on machine
On the machine where you run this code, you need to make sure you can open up enough file descriptors. This can done by 
running:
```
ulimit -n 16384
```
