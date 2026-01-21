# Reproducing the Measurements
In this file you can read how you can reproduce our measurements using the containers and Go code built in according to the instructions of the [README](README.md).

## Measurement Overview
On a high level the measurements were performed with the two servers that were connected to each other and had the following responsibilities:
1. Nameserver server: this was the server that held the zone file and would act as an authoritative nameserver. It does the following:
    - Start nameserver container
    - Wait until this container is ready
    - Start collecting the server statistics
    - Wait for the querier server to signal it is done sending queries while answering said queries
    - Stop collecting the server statistics
    - Prettify the JSON statistics output
2. Querier server: this is the server that fires off the DNS requests to the nameserver server. It does the following:
   - Wait until nameserver container is ready
   - Send queries
   - Signal it is done to nameserver server 

## Machine Configuration
Both of the servers used need some configuration before we are able to perform the measurements. 

On both servers where you run this code, you need to make sure you can open up enough file descriptors and that you can make enough TCP 
connections. This can be done by running:
```bash
    sysctl -w fs.nr_open=33554432
    ulimit -Hn 33554432
    ulimit -Sn 33554432
    sysctl net.ipv4.tcp_fin_timeout=30
    sysctl net.ipv4.tcp_tw_reuse=1
```

Both of the servers in addition to this need to contain the folder `tmp/test-readiness`.

The last configuration needed is to make sure that you can connect from one server to another using `ssh` and that this connection can be done by a script with no user interaction needed. This can be done through these commands (if you have an `ssh` key that the other server trusts):
```bash
    eval "$(ssh-agent -s)"
    ssh-add ~/.ssh/id_rsa # Or some other key filename
```
This needs to be done on both servers.

Now I will explain the scripts that the measurements needed.

## Measurement Scripts
There are several bash and python scripts in this repository which were used to perform the measurements. In the part below, I will try to explain how all of them work.

### stats.sh
The `stats.sh` script takes as input the filename of the output file: the gathered statistics get stored in this file. The script gets the statistics gathered by Podman from a running Podman container every 30 seconds. It then puts the statistics JSON into the output file in a JSON list format. The loop gathering statistics will run indefinitely. You can stop it 
using Ctrl+C, or by sending it an EXIT signal, then it will close the JSON list in the output file correctly.   

To be able to run this script on the server running the nameservers, you need to make sure that the Podman socket is activated. This can be done with this command:
```bash
systemctl --user start podman.socket
```

Afterwards we can find out where we can get this socket using:
```bash
ls $XDG_RUNTIME_DIR/podman/podman.sock
```

The location that this gives as output, you need to put in `stats.sh` in the place where there is currently:
```
/run/user/4002/podman/podman.sock
```
So, you need to change the location that is in the 22nd line of the file. 

### complete_json_file.py
The `complete_json_file.py` takes two command line arguments: 
- the input filename of the JSON statistics file (this file is assumed to be in the same format as the `stats.sh` output) 
- the output filename
With these command line arguments it writes the output of the statistics to a more structured 
JSON format. So you can leave the call to this python code out, if you don't care about JSON being nicely indented. 

### test_synch.sh
This is a script that stops running when the nameserver at a certain IP-address and port has started up (which means it gives a valid DNS response for a domain that we know it serves). The domain it does a DNS request for it defined in `TEST_NAME`, so if you use it, you should change this variable to the one corresponding to a domain you know is in the zone file. 

### run_nameserver_batch_*.sh
These are the files: `run_nameserver_batch_bind.sh`, `run_nameserver_batch_knot.sh`, `run_nameserver_batch_nsd.sh`, and `run_nameserver_batch_pdns.sh`.

Each of these scripts is meant to run on the server that runs the nameservers, and they run all measurements for the 
nameserver implementation specified in the filename. This means that they run the measurement code 10 times for every percentage of TCP (from 0 to 100, so it does that 11 times).

In each of these scripts is the `do_one_measurement` function that holds the action that are performed each measurement. It takes as input the name of the output file and does the following:
1. Start the nameserver in a podman container using the commands lined out in the `Running a Local Nameserver` section of README. The nameserver is started in the background. 
2. Run the `test_synch.sh` script that will stop running when the nameserver we started is ready to answer DNS requests.
3. Start the `stats.sh` script in the background with as input the given output file. 
4. Wait 60 seconds for this script to gather data of the nameserver when it is idle. 
5. Using `ssh` create the file `/tmp/test-readiness/ready` on the other server, to signal that the nameserver is ready to answer DNS requests. 
6. Sleep for 50 minutes and then start checking for whether the `/tmp/test-readiness/ready` file was already created by the other server. This file is made over `ssh` by the querying server to show that it is done sending queries. Once the file is found it is deleted.
7. Sleep another 60 seconds for the `stats.sh` script to gather another 60 seconds of idle data.
8. Kill both the `stats.sh` process and the podman process. 
9. Run `complete_json_file.py` to combine the different measurement outputs into one unified json file

When you want to do a measurement, you need to start your corresponding `run_nameserver_batch_*.sh` script at the same time as the `run_querier_batch_*.sh` script on the other server. These scripts are explained next.

### run_querier_batch_*.sh
These are the files: `run_querier_batch_bind.sh`, `run_querier_batch_knot.sh`, `run_querier_batch_nsd.sh`, and `run_querier_batch_pdns.sh`.

Each of these scripts is meant to run on the server that runs the querier, and they run all measurements corresponding to the specified 
nameserver implementation. This means that they run the measurement code 10 times for every percentage of TCP (from 0 to 100, 
so it does that 11 times). 

In each of these scripts is the `do_one_measurement` function that holds the actions that are performed in each measurement. It takes as input the percentage of queries to retry over TCP, the query input filename (file should be in the format specified in section `Input`) and the output filename. It does the following:
1. Wait until the file `/tmp/test-readiness/ready` is created by the other server, showing that the nameserver has been started. Once the file is created, we remove it immediately.
2. Run the `querier` script with as input file and percentage parameter the ones given as arguments to the function and as output file also the one from the arguments.
3. Once it is done running, use `ssh` to create the `/tmp/test-readiness/ready` file on the other server that shows it we are done sending DNS requests. Note that hence in order to run this script you need to be able to connect with `ssh` to the other server.
4. Sleep for 60 seconds to take into account the 60 seconds of only running `stats.sh` on the second server.

### Measurements Without Retries
The files: `no-retry/querier_no_retry.go`, `run_nameserver_batch_only_tcp.sh`, and `run_querier_only_tcp.sh` all correspond to those needed to perform a slightly different measurement. In this measurement, we don't first send all requests over UDP and then retry a percentage of DNS requests over TCP, but send all requests immediately over TCP. Since this required different Go code we had to write a different `.go` program for this, which is run through these scripts in the same manner as described in the two sections above.


## Jupyter Notebook Analysis
### make_graphs_updated.ipynb
This file is a Jupyter Notebook that holds the python code that we used to analyse our outputs and make the corresponding graphs. All graphs are based on the CPU time output measured by the `stats.sh` script.

The first two graphs that are printed in this notebook are overview graphs that enable you to compare the outcomes between the four different nameserver implementations that we tested. Then the same other less interesting graphs are printed for each of these implementations.

