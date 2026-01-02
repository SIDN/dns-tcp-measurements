#! /bin/bash

kill_jobs() {
    for job in $(jobs -p); do
            kill -s SIGTERM $job > /dev/null 2>&1 || (sleep 10 && kill -9 $job > /dev/null 2>&1 &)

    done
}

trap kill_jobs EXIT

do_one_measurement () {
    # The function gets one arguments: output_filename
    podman run --replace --rm --network=host \
     -v ./nsd/nsd.conf:/etc/nsd/nsd.conf -v ./zones:/dns -v ./nsd/config:/config \
     --name nsd-query querier-nsd:latest nsd -V 2 -d & # '&' makes it run in the background
    local pman_PID=$!
    sleep 240 # wait a bit for the zonefile to load
    date_string=$(date +"%d-%m-%Y_%H:%M:%S")
    echo "[$date_string] Starting statistics"
    # Run the statistics gatherer in the background while discarding its output
    ./stats.sh "$2" &> /dev/null &
    local stats_PID=$!

    # Gather statistics while not doing anything for 60 seconds
    sleep 10
    date_string=$(date +"%d-%m-%Y_%H:%M:%S")
    echo "[$date_string] querier should be starting now"
    # The querier will be running now on the other server
    sleep 600
    date_string=$(date +"%d-%m-%Y_%H:%M:%S")
    echo "[$date_string] Querier should be done now"
    # Once the querier has finished we gather statistics for 60 seconds again
    sleep 10
    date_string=$(date +"%d-%m-%Y_%H:%M:%S")
    echo "[$date_string] done running stats, killing stats.sh and podman"
    # First kill the statistics program
    kill "$stats_PID"
    # Then kill the NSD podman container
    kill "$pman_PID"
    wait "$pman_PID"
    date_string=$(date +"%d-%m-%Y_%H:%M:%S")
    echo "[$date_string] python json stuff starts running now"
    # Call the python script that combines the different outputs into one json file
    python3 complete_json_file.py "$1" "$2"
    date_string=$(date +"%d-%m-%Y_%H:%M:%S")
    echo "[$date_string] python json stuff is done running now"
}


timestamp_str=$(date +"%d-%m-%Y_%H:%M:%S")


# do_one_measurement "test-csv/n1-1h-anon.csv" "stats-output/50_tcp_host_${timestamp_str}.json" "50"
# date_string=$(date +"%d-%m-%Y_%H:%M:%S")
# echo "[$date_string] measurement 50% tcp done"

do_one_measurement "test-csv/n1-1h-anon.csv" "stats-output/50_tcp_host_${timestamp_str}.json"
date_string=$(date +"%d-%m-%Y_%H:%M:%S")
echo "[$date_string] measurement 50% tcp done"
# sleep 10
# do_one_measurement "/home/elmer/ns1data/ns1-1h-anon.csv" "stats-output/100_tcp_host_${timestamp_str}.json" "100"
# echo "measurement 100% tcp done"

