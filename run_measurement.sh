#! /bin/bash

kill_jobs() {
    for job in $(jobs -p); do
            kill -s SIGTERM $job > /dev/null 2>&1 || (sleep 10 && kill -9 $job > /dev/null 2>&1 &)

    done
}

trap kill_jobs EXIT

do_one_measurement () {
    # The function gets three arguments: input_filename, output_filename, percentage_tcp
    podman run --replace --rm --network=host \
     -v ./nsd/nsd.conf:/etc/nsd/nsd.conf -v ./nsd/zones:/dns -v ./nsd/config:/config \
     --name nsd-query querier-nsd:latest nsd -V 2 -d & # '&' makes it run in the background
    local pman_PID=$!
    sleep 2 # wait a bit for the container to start up
    # Run the statistics gatherer in the background while discarding its output
    ./stats.sh "$2" &> /dev/null &
    local stats_PID=$!

    # Gather statistics while not doing anything for 60 seconds
    sleep 60 
    ~/go/bin/querier -f "$1" -p "$3" > "output/querier.out"
    # Once the querier has finished we gather statistics for 60 seconds again
    sleep 60
    # First kill the statistics program
    kill "$stats_PID"
    # Then kill the NSD podman container
    kill "$pman_PID"
    wait "$pman_PID"

    # Call the python script that combines the different outputs into one json file
    python3 complete_json_file.py "$1" "$2" "output/querier.out" 
    
    # Cleanup
    rm output/querier.out
}


timestamp_str=$(date +"%d-%m-%Y_%H:%M:%S")


do_one_measurement "/home/elmer/ns1data/ns1-1h-anon.csv" "stats-output/0_tcp_host_${timestamp_str}.json" "0"
echo "measurement 0% tcp done"
sleep 10
do_one_measurement "/home/elmer/ns1data/ns1-1h-anon.csv" "stats-output/100_tcp_host_${timestamp_str}.json" "100"
echo "measurement 100% tcp done"

