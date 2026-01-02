#! /bin/bash

kill_jobs() {
    for job in $(jobs -p); do
            kill -s SIGTERM $job > /dev/null 2>&1 || (sleep 10 && kill -9 $job > /dev/null 2>&1 &)

    done
}

trap kill_jobs EXIT

do_one_measurement () {
    # The function gets one arguments: output_filename
    eval "$3"
    ./test_synch.sh # this program will exit once the nameserver has loaded
    date_string=$(date +"%d-%m-%Y_%H:%M:%S")
    echo "[$date_string] Starting statistics"
    # Run the statistics gatherer in the background while discarding its output
    ./stats.sh "$2" &> /dev/null &
    local stats_PID=$!

    # Gather statistics while not doing anything for 60 seconds
    sleep 60
    date_string=$(date +"%d-%m-%Y_%H:%M:%S")
    echo "[$date_string] touching the file on eline01"
    ssh eline@145.102.0.141 "touch /tmp/test-readiness/ready"
    date_string=$(date +"%d-%m-%Y_%H:%M:%S")
    echo "[$date_string] querier should be starting now"
    sleep 50m # I guess we can assume it takes at least this many minutes, to make sure it is not too busy waiting
    # Now we wait until the same ready file is created in our tmp file
    while [ ! -f "/tmp/test-readiness/ready" ]; do
        sleep 1 # Every second check whether it was created yet or not
    done
    echo "[$date_string] Querier should be done now"
    rm -f "/tmp/test-readiness/ready" # remove the file so it's not there on the next run
    # Once the querier has finished we gather statistics for 60 seconds again
    sleep 60 
    date_string=$(date +"%d-%m-%Y_%H:%M:%S")
    echo "[$date_string] done running stats, killing stats.sh and podman"
    # First kill the statistics program
    kill "$stats_PID"
    # Then kill the podman container
    podman stop "$4"
    date_string=$(date +"%d-%m-%Y_%H:%M:%S")
    echo "[$date_string] python json stuff starts running now"
    # Call the python script that combines the different outputs into one json file
    python3 complete_json_file.py "$1" "$2"
    date_string=$(date +"%d-%m-%Y_%H:%M:%S")
    echo "[$date_string] python json stuff is done running now"

}

timestamp_str=$(date +"%d-%m-%Y_%H:%M:%S")
measurements_to_do=(1 2 3 4 5 6 7 8 9 10)

NSD_COMMAND="podman run --replace --rm --network=host -v ./nsd/nsd.conf:/etc/nsd/nsd.conf -v ./zones:/dns -v ./nsd/config:/config --name nsd-query querier-nsd:latest nsd -V 2 -d &" 
NSD_NAME="nsd-query"

for measurement in "${measurements_to_do[@]}"
do 
    date_string=$(date +"%d-%m-%Y_%H:%M:%S")
    echo "[$date_string] starting measurement no. $measurement nsd"
    do_one_measurement "test-csv/n1-1h-anon.csv" "stats-output/no-retries/110_tcp_nsd_${timestamp_str}_$measurement.json" "$NSD_COMMAND" "$NSD_NAME"
    date_string=$(date +"%d-%m-%Y_%H:%M:%S")
    echo "[$date_string] measurement no. $measurement nsd done"
    sleep 60
done

KNOT_COMMAND="podman run --replace --rm --network=host -v ./knot/config/knot.conf:/config/knot.conf -v ./zones:/zones --name knot-query docker.io/cznic/knot:latest knotd -v -v &"  
KNOT_NAME="knot-query"

for measurement in "${measurements_to_do[@]}"
do 
    date_string=$(date +"%d-%m-%Y_%H:%M:%S")
    echo "[$date_string] starting measurement no. $measurement knot"
    do_one_measurement "test-csv/n1-1h-anon.csv" "stats-output/no-retries/110_tcp_knot_${timestamp_str}_$measurement.json" "$KNOT_COMMAND" "$KNOT_NAME"
    date_string=$(date +"%d-%m-%Y_%H:%M:%S")
    echo "[$date_string] measurement no. $measurement knot done"
    sleep 60
done


BIND_COMMAND="podman run --replace --rm --network=host  -v ./bind/etc/bind:/etc/bind -v ./zones:/zones -v ./bind/var/cache/bind:/var/cache/bind -v ./bind/var/lib/bind:/var/lib/bind -v ./bind/var/log:/var/log -p 4242:4242/tcp -p 4242:4242/udp -p 127.0.0.1:1953:953/tcp --name bind-echte-test internetsystemsconsortium/bind9:9.20 -u bind -c /etc/bind/named.conf -f &"
BIND_NAME="bind-echte-test"

for measurement in "${measurements_to_do[@]}"
do 
    date_string=$(date +"%d-%m-%Y_%H:%M:%S")
    echo "[$date_string] starting measurement no. $measurement bind"
    do_one_measurement "test-csv/n1-1h-anon.csv" "stats-output/no-retries/110_tcp_bind_${timestamp_str}_$measurement.json" "$BIND_COMMAND" "$BIND_NAME"
    date_string=$(date +"%d-%m-%Y_%H:%M:%S")
    echo "[$date_string] measurement no. $measurement bind done"
    sleep 60
done

PDNS_COMMAND="podman run --replace --rm --network=host -d -p 4242:4242/udp -p 4242:4242/tcp --name=pqc_pdns pdns-test"
PDNS_NAME="pqc_pdns"

for measurement in "${measurements_to_do[@]}"
do 
    date_string=$(date +"%d-%m-%Y_%H:%M:%S")
    echo "[$date_string] starting measurement no. $measurement pdns"
    do_one_measurement "test-csv/n1-1h-anon.csv" "stats-output/no-retries/110_tcp_pdns_${timestamp_str}_$measurement.json" "$PDNS_COMMAND" "$PDNS_NAME"
    date_string=$(date +"%d-%m-%Y_%H:%M:%S")
    echo "[$date_string] measurement no. $measurement pdns done"
    sleep 60
done