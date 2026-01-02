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
     -v ./knot/config/knot.conf:/config/knot.conf -v ./zones:/zones \
     --name knot-query docker.io/cznic/knot:3.5 knotd -v -v & # '&' makes it run in the background
    local pman_PID=$!
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
    ssh eline@eline01.s6s.nl "touch /tmp/test-readiness/ready"
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
    # Then kill the Knot podman container
    kill "$pman_PID"
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

measurements_to_do=(1 2 3 4 5 6 7 8 9 10)

for measurement in "${measurements_to_do[@]}"
do 
    date_string=$(date +"%d-%m-%Y_%H:%M:%S")
    echo "[$date_string] starting measurement no. $measurement 0% tcp"
    do_one_measurement "test-csv/n1-1h-anon.csv" "stats-output/knot/0_tcp_host_${timestamp_str}_$measurement.json"
    date_string=$(date +"%d-%m-%Y_%H:%M:%S")
    echo "[$date_string] measurement no. $measurement 0% tcp done"
    sleep 60
done

for measurement in "${measurements_to_do[@]}"
do 
    date_string=$(date +"%d-%m-%Y_%H:%M:%S")
    echo "[$date_string] starting measurement no. $measurement 10% tcp"
    do_one_measurement "test-csv/n1-1h-anon.csv" "stats-output/knot/10_tcp_host_${timestamp_str}_$measurement.json"
    date_string=$(date +"%d-%m-%Y_%H:%M:%S")
    echo "[$date_string] measurement no. $measurement 10% tcp done"
    sleep 60
done

for measurement in "${measurements_to_do[@]}"
do 
    date_string=$(date +"%d-%m-%Y_%H:%M:%S")
    echo "[$date_string] starting measurement no. $measurement 20% tcp"
    do_one_measurement "test-csv/n1-1h-anon.csv" "stats-output/knot/20_tcp_host_${timestamp_str}_$measurement.json"
    date_string=$(date +"%d-%m-%Y_%H:%M:%S")
    echo "[$date_string] measurement no. $measurement 20% tcp done"
    sleep 60
done

for measurement in "${measurements_to_do[@]}"
do 
    date_string=$(date +"%d-%m-%Y_%H:%M:%S")
    echo "[$date_string] starting measurement no. $measurement 30% tcp"
    do_one_measurement "test-csv/n1-1h-anon.csv" "stats-output/knot/30_tcp_host_${timestamp_str}_$measurement.json"
    date_string=$(date +"%d-%m-%Y_%H:%M:%S")
    echo "[$date_string] measurement no. $measurement 30% tcp done"
    sleep 60
done

for measurement in "${measurements_to_do[@]}"
do 
    date_string=$(date +"%d-%m-%Y_%H:%M:%S")
    echo "[$date_string] starting measurement no. $measurement 40% tcp"
    do_one_measurement "test-csv/n1-1h-anon.csv" "stats-output/knot/40_tcp_host_${timestamp_str}_$measurement.json"
    date_string=$(date +"%d-%m-%Y_%H:%M:%S")
    echo "[$date_string] measurement no. $measurement 40% tcp done"
    sleep 60
done

for measurement in "${measurements_to_do[@]}"
do 
    date_string=$(date +"%d-%m-%Y_%H:%M:%S")
    echo "[$date_string] starting measurement no. $measurement 50% tcp"
    do_one_measurement "test-csv/n1-1h-anon.csv" "stats-output/knot/50_tcp_host_${timestamp_str}_$measurement.json"
    date_string=$(date +"%d-%m-%Y_%H:%M:%S")
    echo "[$date_string] measurement no. $measurement 50% tcp done"
    sleep 60
done

for measurement in "${measurements_to_do[@]}"
do 
    date_string=$(date +"%d-%m-%Y_%H:%M:%S")
    echo "[$date_string] starting measurement no. $measurement 60% tcp"
    do_one_measurement "test-csv/n1-1h-anon.csv" "stats-output/knot/60_tcp_host_${timestamp_str}_$measurement.json"
    date_string=$(date +"%d-%m-%Y_%H:%M:%S")
    echo "[$date_string] measurement no. $measurement 60% tcp done"
    sleep 60
done

for measurement in "${measurements_to_do[@]}"
do 
    date_string=$(date +"%d-%m-%Y_%H:%M:%S")
    echo "[$date_string] starting measurement no. $measurement 70% tcp"
    do_one_measurement "test-csv/n1-1h-anon.csv" "stats-output/knot/70_tcp_host_${timestamp_str}_$measurement.json"
    date_string=$(date +"%d-%m-%Y_%H:%M:%S")
    echo "[$date_string] measurement no. $measurement 70% tcp done"
    sleep 60
done

for measurement in "${measurements_to_do[@]}"
do 
    date_string=$(date +"%d-%m-%Y_%H:%M:%S")
    echo "[$date_string] starting measurement no. $measurement 80% tcp"
    do_one_measurement "test-csv/n1-1h-anon.csv" "stats-output/knot/80_tcp_host_${timestamp_str}_$measurement.json"
    date_string=$(date +"%d-%m-%Y_%H:%M:%S")
    echo "[$date_string] measurement no. $measurement 80% tcp done"
    sleep 60
done

for measurement in "${measurements_to_do[@]}"
do 
    date_string=$(date +"%d-%m-%Y_%H:%M:%S")
    echo "[$date_string] starting measurement no. $measurement 90% tcp"
    do_one_measurement "test-csv/n1-1h-anon.csv" "stats-output/knot/90_tcp_host_${timestamp_str}_$measurement.json"
    date_string=$(date +"%d-%m-%Y_%H:%M:%S")
    echo "[$date_string] measurement no. $measurement 90% tcp done"
    sleep 60
done

for measurement in "${measurements_to_do[@]}"
do 
    date_string=$(date +"%d-%m-%Y_%H:%M:%S")
    echo "[$date_string] starting measurement no. $measurement 100% tcp"
    do_one_measurement "test-csv/n1-1h-anon.csv" "stats-output/knot/100_tcp_host_${timestamp_str}_$measurement.json"
    date_string=$(date +"%d-%m-%Y_%H:%M:%S")
    echo "[$date_string] measurement no. $measurement 100% tcp done"
    sleep 60
done
