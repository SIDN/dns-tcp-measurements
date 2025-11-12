# DNS querier
Querier for replaying DNS queries against an authoritative nameserver. The replay of these queries is done by trying to recreate the timing of each message.

## Input
The dataformat of the input file should be as follows:
```
0.001,example.nl,AAAA,UDP,0
```
This is in the following order: `time,qname,qtype,prot,DO`. Where `offset` is the time in seconds relative to the start of the replay at which the query should be sent. `DO` is the value of the DO-bit in the query. The other types are pretty straightforward, so if it should use UDP or TCP, what domain name it should request and what type of 
record it should request. 

## Build and usage
Build with `go install gitlab.sidnlabs.nl/eline.stehouwer/querier`, then run with `~/go/bin/querier`. \
Before running the code, it is important that you also have a nameserver running that you run the code against, 
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
podman run --rm --network=host -it -v ./nsd.conf:/etc/nsd/nsd.conf -v ./zones:/dns -v ./config:/config -p 4242:53/tcp -p 4242:53/udp --name nsd-query querier-nsd:latest nsd -V 2 -d 
```
In this command we specify that the network used must be `host`. This is because the `slirp4netns` network seems to be slower and is less able to handle many TCP requests compared to the `pasta` and `host` network. The `host` network is able to handle TCP requests a lot faster than the `pasta` network. \
After you have this running in a container you can run the `querier` as explained above. 

## Needed configuration on machine
On the machine where you run this code, you need to make sure you can open up enough file descriptors. This can done by 
running:
```
ulimit -n 16384
```

# Measurements
## stats.sh
The `stats.sh` script takes as input the output filename, in which the gathered statistics get stroed. The script gets the statistics gathered by podman from a running podman container every 10 seconds. It then puts the statistics JSON into the output file in a JSON list format. The loop gathering statistics will run indefinitely, you can stop it 
using Ctl+C, or by sending it an EXIT signal, then it will close the JSON list in the output file correctly.   

## complete_json_file.py
The `complete_json_file.py` takes three command line arguments: the input filename of the JSON statistics file (this file is assumed to be in the same format as the `stats.sh` output), the output filename and the querier output file (this is the file that the `querier` go program has written it's output to). With these command line arguments it first 
parses the `querier` output and then it writes these together with the JSON statistics into the given output file in a structured JSON format. 

## run_measurement.sh
With the `run_measurement.sh` script, a measurement using the DNS querier can be run. This measurement script has a 
`do_one_measurement` function, that takes the following inputs:
- The `.csv` input file formatted as specified in section `Input`. 
- The output file that the resulting JSON should be stored in
- The percentage of queries in string format that need to retry over TCP

With these inputs it then does the following in specified order:
1. Start the nameserver in a podman container in the background using the command as specified in the previous `Building local nameserver` section.
2. Start the `stats.sh` script in the background with as output file the output file given as argument
3. Sleep for 60 seconds, such that the `stats.sh` script can gather measurements of an idle nameserver
4. Run the `querier` script with as input file and percentage parameter the ones given as arguments to the function and as output file: `output/querier.out`. 
5. Sleep for 60 seconds to again have `stats.sh` gather measurements of an idle nameserver
6. Kill the statistics program
7. Kill the podman container. 
8. Run `complete_json_file.py` to combine the different measurement outputs into one unified json file
9. Remove the `output/querier.out` file.

Right now running the script will run the `do_one_measurement` function once, using `/home/elmer/ns1data/ns1-1h-anon.csv` as input setting the TCP retry percentage to 50. 

There is also another script `run_measurement_batch.sh`. I use this script to run all the measurements that I want to run in one go. This means for every percentage of TCP it runs the `do_one_measurement` function 10 times. 