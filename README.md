# DNS Query Replayer
## Overview
Query replayer for replaying DNS queries against an authoritative nameserver. The replay of these queries is done by trying to recreate the timing of each message. 

In addition to this query replayer, the repository also contains code to run it against four different nameserver implementations and during this run measure the CPU usage used by the nameserver while it processes and answers the queries it receives. 

This replayer was made with the goal of using it in measurements to estimate the impact of an increased percentage of TCP retries on DNS. This repository also contains script to perform these measurements.

## Quick Start
1. Build the querier:
```bash
   go install gitlab.sidnlabs.nl/eline.stehouwer/querier
```
2. Start a local nameserver, one option is via containers. So if you want to take NSD:
```bash
    cd nsd
    podman build -f Dockerfile --tag=querier-nsd:latest
    podman run --rm --network=host -it \
      -v ./nsd.conf:/etc/nsd/nsd.conf \
      -v ./zones:/dns \
      --name nsd-query querier-nsd:latest nsd -V 2 -d
```
3. Run the query replay tool:
```bash
    ~/go/bin/querier -f test-csv/test_file_structure.csv -s 127.0.0.1:4242
```

## Using the DNS Query Replayer
### Input Format
For the replayer to be able to send queries, it needs a CSV input file as input. The data format of this file should be as follows:
```
0.001,example.nl,AAAA,UDP,0
```
This is in the following order: `time,qname,qtype,prot,DO`.

- `time`: offset in seconds relative to the start of the replay
- `qname`: domain name to query
- `qtype`: DNS record type (e.g., A, AAAA)
- `prot`: UDP or TCP
- `DO`: value of the DNSSEC DO bit (0 or 1) 

### Build and Run
Build with `go install gitlab.sidnlabs.nl/eline.stehouwer/querier`, then run with `~/go/bin/querier`. \
Before running the code, you must also have an authoritative DNS nameserver running that you run the code against, if you want to test it against a local nameserver.  

By default, the querier sends all queries over UDP and retries a configurable percentage over TCP.

### Command-Line Arguments
It can take three command line arguments:
- `-f <file>`  
  CSV input file with queries.  
  Default: `test-csv/test_file_structure.csv`
- `-p <percentage>`  
  Percentage of queries retried over TCP (0–100).  
  Default: `100`
- `-s <addr>`  
  Nameserver address in `[IP]:[port]` format.  
  Default: `127.0.0.1:4242`

### Output
The querier writes to an output file containing:
- Number of succesful responses
- Total DNS requests sent (including and excluding the extra TCP retries)
- Execution time
- Rcodes of the responses
- Any errors that occurred during the run

These can mainly be used to visually inspect whether anything went wrong. 

## Running a Local Nameserver (Containers)
If you want to replay your DNS queries to a local nameserver with a zone that you configure, you can use these instructions to set that host up. We used Podman container to run all of our local authoritative nameservers.

### Common Notes
In all our Podman run commands, we specify that the network used must be `host`. This is because the `slirp4netns` network is to be slower and is less able to handle many TCP requests compared to the `pasta` and `host` network. The `host` network can handle TCP requests a lot faster than the `pasta` network. 

For each of the nameserver containers you need to put the zone files and the other needed configuration files in the right places, make the configuration files point to the right zone files and mount them at the right places when you run the container. 

If you have kept the config file the same, then you can test whether a container is up in the air and working with:
```bash
    dig @localhost -p 4242 [SOME-DOMAIN-NAME-HERE]
```

### NSD
In this repository, under the `nsd` folder, we have put a simple Dockerfile that installs a `nsd` nameserver and the corresponding 
file structure needed to set this nameserver up. To get this nameserver up and running, you only need to provide 
the right files and build the container as follows:
```
podman build -f Dockerfile --tag=querier-nsd:latest
```
Then afterwards you can run you the `nsd` nameserver in a container with the following command:
```
podman run --rm --network=host -it -v ./nsd.conf:/etc/nsd/nsd.conf -v ./zones:/dns --name nsd-query querier-nsd:latest nsd -V 2 -d 
```

### BIND
To get the BIND implementation up and running, we use the official BIND9 container published by the ISC. So, to use it, we first need to pull it from the container hub, which can be done as follows:
```
podman pull docker.io/internetsystemsconsortium/bind9:9.20
```
The configuration file and folder structure needed for the BIND nameserver is located in the `bind` folder (with the configuration file located in `bind/etc/bind`).

Once the configuration is ready, we can start the container like this:
```
podman run --replace --rm --network=host  -v ./bind/etc/bind:/etc/bind -v ./zones:/zones -v ./bind/var/cache/bind:/var/cache/bind -v ./bind/var/lib/bind:/var/lib/bind -v ./bind/var/log:/var/log --name bind-echte-test internetsystemsconsortium/bind9:9.20 -u bind -c /etc/bind/named.conf -f
```

### Knot

Just like with BIND for Knot, we use the published official container. So, first we have to pull this again:
```
podman pull docker.io/cznic/knot:v3.5.2
```

In the `knot` folder, the needed folder structure is given. Once the configuration is as wanted, we can start the container:
```
podman run --replace --rm --network=host -v ./knot/config/knot.conf:/config/knot.conf -v ./zones:/zones --name knot-query docker.io/cznic/knot:v3.5.2 knotd -v -v
```
Where we again mount the folders we need in the container using the `-v` option and use the `host` network option. 


### PowerDNS

Although we also wanted to use a container on Docker Hub for the PowerDNS implementation, these did not work as we expected, so we had to do something else. We instead built our own container from a published PowerDNS release. So we build that using the Dockerfile in the `pdns` folder and the following command:

```
podman build -f pdns/Dockerfile --tag=auth-powerdns
```

To then load our zone file into the container and make sure the signing is set up correctly we made a second container. This is done in the Dockerfile inside the `pdns/4o` folder. So first we needed to build this container:
```
podman build -f pdns/4o/Dockerfile --tag=pdns-test:latest 
```
To change this container to use a different zone file, you have to change both the Dockerfile and the `pdns.conf` file. 
The zone we used was already presigned, so we could just do `set-presigned` and then PowerDNS did not have to do any signing.

After building the container we can run it without mounting any folders, since these were already added in the Dockerfile we built the container with. 
We can run it like this:
```
podman run --replace --rm --network=host -d --name=pqc_pdns pdns-test
```

## Reproducing the Measurements
You can read how the measurements were performed and how you can reproduce them in the file [MEASUREMENTS.md](MEASUREMENTS.md). We separated this to keep the README a bit more compact. 
