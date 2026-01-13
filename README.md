# DNS querier
Querier for replaying DNS queries against an authoritative nameserver. The replay of these queries is done by trying to recreate the timing of each message. 

In addition to this querier, the repository also contains code to run this querier against four different nameserver implementations and during this run measure the CPU usage used by the nameserver while it processes and answers the queries it receives. 

## Input
For the querier to be able to send queries it needs an input file as input. The data format of this file should be as follows:
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
- `-s [server address]`: with the `-s` option you can pass the address of the name server that you want to send the queries 
to. The address is to be formatted in the following way: `[IP address]:[port]`. The default value is `127.0.0.1:4242`.

### Building and running local nameserver
As already mentioned, we have run our querier against 4 different nameserver implementation, and for each of them we have the 
instructions to set them up for the experiment below. For portability, we run each of these implementations from a container using `podman`.
#### NSD
In this repository, under the `nsd` folder, we have put a simple Dockerfile that installs a `nsd` nameserver and the corresponding 
file structure needed to set this nameserver up. To get this nameserver up and running you only need to provide 
the zone files yourself (and edit the .conf file to point to these zone files) and build the container as follows:
```
podman build -f Dockerfile --tag=querier-nsd:latest
```
Then afterwards you can run you the `nsd` nameserver in a container with the following command (provided you have the 
right file structure)
```
podman run --rm --network=host -it -v ./nsd.conf:/etc/nsd/nsd.conf -v ./zones:/dns --name nsd-query querier-nsd:latest nsd -V 2 -d 
```
In this command we specify that the network used must be `host`. This is because the `slirp4netns` network seems to be slower and is less able to handle many TCP requests compared to the `pasta` and `host` network. The `host` network is able to handle TCP requests a lot faster than the `pasta` network. \
In addition to this we put our `./zones` folder, which contains our zone file, at the `/dns` folder in the image, since that is where we said it would be in our configuration file. \
After you have this running in a container you can send DNS requests to it. 

#### BIND
To get the BIND implementation up and running we use the official BIND9 container published by the ISC. So, in order to use it we first need to pull 
it from the container hub, which can be done as follows:
```
podman pull docker.io/internetsystemsconsortium/bind9:9.20
```
The configuration file and folder structure needed for the BIND nameserver is located in the bind folder (with the configuration file located in `bind/etc/bind`). The configuration file needs to be changed to the settings that you prefer. The image also needs to be able to access the zone file it is going to server, and you need to change the configuration file to point to the right path in the image where you will put the zone file. 

Once the configuration is ready we can start the container like this:
```
podman run --replace --rm --network=host  -v ./bind/etc/bind:/etc/bind -v ./zones:/zones -v ./bind/var/cache/bind:/var/cache/bind -v ./bind/var/lib/bind:/var/lib/bind -v ./bind/var/log:/var/log --name bind-echte-test internetsystemsconsortium/bind9:9.20 -u bind -c /etc/bind/named.conf -f
```
In this command, we again use the host-network, and we also put the `./zones` folder with our zone file in a folder `/zones` in the image. In addition to this, we couple the folder structure we made in the `bind` folder with the corresponding folders in the image. To make sure the container has the correct rights we specify that the user is `bind`. \
This should give an up and running BIND nameserver that we can send DNS requests to.

#### Knot
Just like with BIND for Knot we use the published official container. So, first we have to pull this again:
```
podman pull docker.io/cznic/knot:v3.5.2
```

In the `knot` folder the needed folder structure is given. To run the container, you need to change the `knot/knot.conf` file to your liking, especially the location of the zone file. 

Once the configuration is as wanted, we can start the container:
```
podman run --replace --rm --network=host -v ./knot/config/knot.conf:/config/knot.conf -v ./zones:/zones --name knot-query docker.io/cznic/knot:v3.5.2 knotd -v -v
```
Where we again mount the folders we need in the container using the `-v` option and use the `host` network option. 

#### PowerDNS
Although we also wanted to use a container on the docker hub for the PowerDNS implementation, these did not work as we expected, so we had to do something else. We instead used the PowerDNS container as published by SIDN Labs that is patched with PQC algorithms, which can be found on https://github.com/sidn/pqc-auth-powerdns/pkgs/container/pqc-auth-powerdns. 

To then actually use this container with our zone file, we needed to build a new container from that. This is done in the Dockerfile inside the `pdns/4o` folder. So first we needed to build this container:
```
podman build -f pdns/4o/Dockerfile --tag=pdns-test:latest 
```
To change this container to use a different zone file you need to change both the Dockerfile and the `pdns.conf` file. 
The zone we used was already presigned, so we could just do `set-presigned` and then PowerDNS did not have to do any signing. 

After building the container we can run it without mounting any folders, since these were already added in the Dockerfile we built the container with. 
We can run it like this:
```
podman run --replace --rm --network=host -d --name=pqc_pdns pdns-test
```
This has the nameserver up and running.


# Measurements
The measurement setup that we had during our measurements was: two servers. One that runs the Go querier code and sends DNS requests with a predefined order and timing to the other server and waits for the response. The other that runs one of the four nameserver implementations with the zone file corresponding to the queries that it gets, so it does as a normal nameserver does: wait for requests and respond to them with a DNS response.

In this part I will first explain the needed configuration on the servers, and then the different scripts that are needed during the measurement process. 

## Needed configuration on machine
If you want to perform the same measurements that we did, it is important that you set up the machines you run them on with the same settings. 
On both servers where you run this code, you need to make sure you can open up enough file descriptors and that you are able to make enough TCP 
connections. This can be done by running:
```
sysctl -w fs.nr_open=33554432
ulimit -Hn 33554432
ulimit -Sn 33554432
sysctl net.ipv4.tcp_fin_timeout=30
sysctl net.ipv4.tcp_tw_reuse=1
```

Both of the servers also need to contain the folder `tmp/test-readiness`.

## stats.sh
The `stats.sh` script takes as input the output filename, in which the gathered statistics get stored. The script gets the statistics gathered by podman from a running podman container every 30 seconds. It then puts the statistics JSON into the output file in a JSON list format. The loop gathering statistics will run indefinitely, you can stop it 
using Ctrl+C, or by sending it an EXIT signal, then it will close the JSON list in the output file correctly.   

To be able to run this script on the server running the nameservers, you need to make sure that the podman socket is activated. This can be done with this command:
```
systemctl --user start podman.socket
```
Afterwards we can find out where we can get this socket using:
```
ls $XDG_RUNTIME_DIR/podman/podman.sock
```
The location that this gives as output, you need to put in `stats.sh` in the place where there is currently:
```
/run/user/4002/podman/podman.sock
```
So, in the 22nd line of the file. Afterwards the code will work as expected. 

## complete_json_file.py
The `complete_json_file.py` takes two command line arguments: the input filename of the JSON statistics file (this file is assumed to be in the same format as the `stats.sh` output) and the output filename. With these command line arguments it writes the output of the statistics to a more structured 
JSON format. 

## test_synch.sh
This is a script that stops running when the nameserver at a certain IP-address and port has started up (which means it gives a valid DNS response for a domain that we know it serves). The domain it does a DNS request for it defined in `TEST_NAME`, so if you use it, you should change this variable to the one corresponding to a domain you know is in the zone file. 

## run_nameserver_batch_*.sh
These are the files: `run_nameserver_batch_bind.sh`, `run_nameserver_batch_knot.sh`, `run_nameserver_batch_nsd.sh` and `run_nameserver_batch_pdns.sh`.

Each of these scripts is meant to run on the server that runs the nameservers, and they run all measurements for the specified 
nameserver implementation. This means that they run the measurement code 10 times for every percentage of TCP (from 0 to 100, 
so it does that 11 times). 

In each of these scripts is the `do_one_measurement` function that holds the action that are performed each measurement. It takes as input the name of the output file and does the following:
1. Start the nameserver in a podman container using the commands lined out in the `Building and running a local nameserver`. The nameserver is started in the background. 
2. Run the `test_synch.sh` script that will stop running when the nameserver we started is ready to answer DNS requests.
3. Start the `stats.sh` script in the background with as input the given output file. 
4. Wait 60 seconds for this script to gather data of the nameserver when it is idle. 
5. Using `ssh` create the file `/tmp/test-readiness/ready` on the other server, to signal that the nameserver is ready to answer DNS requests. Note that hence in order to run this script you need to be able to connect with `ssh` to the other server.
6. Sleep for 50 minutes and then start checking for whether the `/tmp/test-readiness/ready` file was already created by the other server. This file is made over `ssh` by the querying server to show that it is done sending queries. Once the file is found it is deleted.
7. Sleep another 60 seconds for the `stats.sh` script to gather another 60 seconds of idle data.
8. Kill both the `stats.sh` process and the podman process. 
9. Run `complete_json_file.py` to combine the different measurement outputs into one unified json file

When you want to run a measurement you need to start your `run_nameserver_batch_*.sh` script at the same time as the `run_querier_batch_*.sh` script on the other server. This script is explained next.

## run_querier_batch_*.sh
These are the files: `run_querier_batch_bind.sh`, `run_querier_batch_knot.sh`, `run_querier_batch_nsd.sh` and `run_querier_batch_pdns.sh`.

Each of these scripts is meant to run on the server that runs the querier, and they run all measurements corresponding to the specified 
nameserver implementation. This means that they run the measurement code 10 times for every percentage of TCP (from 0 to 100, 
so it does that 11 times). 

In each of these scripts is the `do_one_measurement` function that holds the action that are performed each measurement. It takes as input the percentage of queries to retry over TCP, the query input filename (file should be in the format specified in section `Input`) and the output filename. It does the following:
1. Wait until the file `/tmp/test-readiness/ready` is created by the other server, showing that the nameserver has been started. Once the file is created we remove it immediately.
2. Run the `querier` script with as input file and percentage parameter the ones given as arguments to the function and as output file also the one from the arguments.
3. Once it is done running, use `ssh` to create the `/tmp/test-readiness/ready` file on the other server that shows it we are done sending DNS requests. Note that hence in order to run this script you need to be able to connect with `ssh` to the other server.
4. Sleep for 60 seconds to take into account the 60 seconds of only running `stats.sh` on the second server.

## Measurements without retries
The files: `no-retry/querier_no_retry.go`, `run_nameserver_batch_only_tcp.sh`, and `run_querier_only_tcp.sh` all correspond to those needed to perform a slightly different measurement, where we don't first send all requests over UDP and then retry a percentage of DNS requests over TCP, but send all requests immediately over TCP. Since this required different Go code we had to write a different `.go` program for this, which is run through these scripts in the same manner as described in the two sections above.

## analysis/make_graphs_updated.ipynb
This file is a Jupyter Notebook that holds the python code that we used to analyse our outputs and make the corresponding graphs.
