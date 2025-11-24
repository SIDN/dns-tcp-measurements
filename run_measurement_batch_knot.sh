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
     -v ./knot/config/knot.conf:/config/knot.conf -v ./zones:/zones \
     --name knot-query docker.io/cznic/knot:latest knotd -v -v & # '&' makes it run in the background
    local pman_PID=$!
    sleep 180 # wait a bit for the zonefile to load
    date_string=$(date +"%d-%m-%Y_%H:%M:%S")
    echo "[$date_string] Starting statistics"
    # Run the statistics gatherer in the background while discarding its output
    ./stats.sh "$2" &> /dev/null &
    local stats_PID=$!

    # Gather statistics while not doing anything for 60 seconds
    sleep 60
    date_string=$(date +"%d-%m-%Y_%H:%M:%S")
    echo "[$date_string] Starting querier" 
    ~/go/bin/querier -f "$1" -p "$3" > "output/querier.out"
    date_string=$(date +"%d-%m-%Y_%H:%M:%S")
    echo "[$date_string]Querier done"
    # Once the querier has finished we gather statistics for 60 seconds again
    sleep 60
    # First kill the statistics program
    kill "$stats_PID"
    # Then kill the NSD podman container
    kill "$pman_PID"
    kill "$pman_PID"
    wait "$pman_PID"

    # Call the python script that combines the different outputs into one json file
    python3 complete_json_file.py "$1" "$2" "output/querier.out" 
    
    # Cleanup
    rm output/querier.out
}


timestamp_str=$(date +"%d-%m-%Y_%H:%M:%S")

measurements_to_do=(1 2 3 4 5 6 7 8 9 10)

for measurement in "${measurements_to_do[@]}"
do 
    date_string=$(date +"%d-%m-%Y_%H:%M:%S")
    echo "[$date_string] starting measurement no. $measurement 0% tcp"
    do_one_measurement "test-csv/n1-1h-anon.csv" "stats-output/knot/0_tcp_host_${timestamp_str}_$measurement.json" "0"
    date_string=$(date +"%d-%m-%Y_%H:%M:%S")
    echo "[$date_string] measurement no. $measurement 0% tcp done"
    sleep 60
done

for measurement in "${measurements_to_do[@]}"
do 
    date_string=$(date +"%d-%m-%Y_%H:%M:%S")
    echo "[$date_string] starting measurement no. $measurement 10% tcp"
    do_one_measurement "test-csv/n1-1h-anon.csv" "stats-output/knot/10_tcp_host_${timestamp_str}_$measurement.json" "10"
    date_string=$(date +"%d-%m-%Y_%H:%M:%S")
    echo "[$date_string] measurement no. $measurement 10% tcp done"
    sleep 60
done

for measurement in "${measurements_to_do[@]}"
do 
    date_string=$(date +"%d-%m-%Y_%H:%M:%S")
    echo "[$date_string] starting measurement no. $measurement 20% tcp"
    do_one_measurement "test-csv/n1-1h-anon.csv" "stats-output/knot/20_tcp_host_${timestamp_str}_$measurement.json" "20"
    date_string=$(date +"%d-%m-%Y_%H:%M:%S")
    echo "[$date_string] measurement no. $measurement 20% tcp done"
    sleep 60
done

for measurement in "${measurements_to_do[@]}"
do 
    date_string=$(date +"%d-%m-%Y_%H:%M:%S")
    echo "[$date_string] starting measurement no. $measurement 30% tcp"
    do_one_measurement "test-csv/n1-1h-anon.csv" "stats-output/knot/30_tcp_host_${timestamp_str}_$measurement.json" "30"
    date_string=$(date +"%d-%m-%Y_%H:%M:%S")
    echo "[$date_string] measurement no. $measurement 30% tcp done"
    sleep 60
done

for measurement in "${measurements_to_do[@]}"
do 
    date_string=$(date +"%d-%m-%Y_%H:%M:%S")
    echo "[$date_string] starting measurement no. $measurement 40% tcp"
    do_one_measurement "test-csv/n1-1h-anon.csv" "stats-output/knot/40_tcp_host_${timestamp_str}_$measurement.json" "40"
    date_string=$(date +"%d-%m-%Y_%H:%M:%S")
    echo "[$date_string] measurement no. $measurement 40% tcp done"
    sleep 60
done

for measurement in "${measurements_to_do[@]}"
do 
    date_string=$(date +"%d-%m-%Y_%H:%M:%S")
    echo "[$date_string] starting measurement no. $measurement 50% tcp"
    do_one_measurement "test-csv/n1-1h-anon.csv" "stats-output/knot/50_tcp_host_${timestamp_str}_$measurement.json" "50"
    date_string=$(date +"%d-%m-%Y_%H:%M:%S")
    echo "[$date_string] measurement no. $measurement 50% tcp done"
    sleep 60
done

for measurement in "${measurements_to_do[@]}"
do 
    date_string=$(date +"%d-%m-%Y_%H:%M:%S")
    echo "[$date_string] starting measurement no. $measurement 60% tcp"
    do_one_measurement "test-csv/n1-1h-anon.csv" "stats-output/knot/60_tcp_host_${timestamp_str}_$measurement.json" "60"
    date_string=$(date +"%d-%m-%Y_%H:%M:%S")
    echo "[$date_string] measurement no. $measurement 60% tcp done"
    sleep 60
done

for measurement in "${measurements_to_do[@]}"
do 
    date_string=$(date +"%d-%m-%Y_%H:%M:%S")
    echo "[$date_string] starting measurement no. $measurement 70% tcp"
    do_one_measurement "test-csv/n1-1h-anon.csv" "stats-output/knot/70_tcp_host_${timestamp_str}_$measurement.json" "70"
    date_string=$(date +"%d-%m-%Y_%H:%M:%S")
    echo "[$date_string] measurement no. $measurement 70% tcp done"
    sleep 60
done

for measurement in "${measurements_to_do[@]}"
do 
    date_string=$(date +"%d-%m-%Y_%H:%M:%S")
    echo "[$date_string] starting measurement no. $measurement 80% tcp"
    do_one_measurement "test-csv/n1-1h-anon.csv" "stats-output/knot/80_tcp_host_${timestamp_str}_$measurement.json" "80"
    date_string=$(date +"%d-%m-%Y_%H:%M:%S")
    echo "[$date_string] measurement no. $measurement 80% tcp done"
    sleep 60
done

for measurement in "${measurements_to_do[@]}"
do 
    date_string=$(date +"%d-%m-%Y_%H:%M:%S")
    echo "[$date_string] starting measurement no. $measurement 90% tcp"
    do_one_measurement "test-csv/n1-1h-anon.csv" "stats-output/knot/90_tcp_host_${timestamp_str}_$measurement.json" "90"
    date_string=$(date +"%d-%m-%Y_%H:%M:%S")
    echo "[$date_string] measurement no. $measurement 90% tcp done"
    sleep 60
done

for measurement in "${measurements_to_do[@]}"
do 
    date_string=$(date +"%d-%m-%Y_%H:%M:%S")
    echo "[$date_string] starting measurement no. $measurement 100% tcp"
    do_one_measurement "test-csv/n1-1h-anon.csv" "stats-output/knot/100_tcp_host_${timestamp_str}_$measurement.json" "100"
    date_string=$(date +"%d-%m-%Y_%H:%M:%S")
    echo "[$date_string] measurement no. $measurement 100% tcp done"
    sleep 60
done

# date_string=$(date +"%d-%m-%Y_%H:%M:%S")
# echo "[$date_string] starting measurement no. 1 0% tcp"
# do_one_measurement "test-csv/n1-1h-anon.csv" "stats-output/0_tcp_host_${timestamp_str}_1.json" "0"
# date_string=$(date +"%d-%m-%Y_%H:%M:%S")
# echo "[$date_string] measurement no. 1 0% tcp done"

# sleep 60

# date_string=$(date +"%d-%m-%Y_%H:%M:%S")
# echo "[$date_string] starting measurement no. 2 0% tcp"
# do_one_measurement "test-csv/n1-1h-anon.csv" "stats-output/0_tcp_host_${timestamp_str}_2.json" "0"
# date_string=$(date +"%d-%m-%Y_%H:%M:%S")
# echo "[$date_string] measurement no. 2 0% tcp done"

# sleep 60

# date_string=$(date +"%d-%m-%Y_%H:%M:%S")
# echo "[$date_string] starting measurement no. 3 0% tcp"
# do_one_measurement "test-csv/n1-1h-anon.csv" "stats-output/0_tcp_host_${timestamp_str}_3.json" "0"
# date_string=$(date +"%d-%m-%Y_%H:%M:%S")
# echo "[$date_string] measurement no. 3 0% tcp done"

# sleep 60

# date_string=$(date +"%d-%m-%Y_%H:%M:%S")
# echo "[$date_string] starting measurement no. 4 0% tcp"
# do_one_measurement "test-csv/n1-1h-anon.csv" "stats-output/0_tcp_host_${timestamp_str}_4.json" "0"
# date_string=$(date +"%d-%m-%Y_%H:%M:%S")
# echo "[$date_string] measurement no. 4 0% tcp done"

# sleep 60

# date_string=$(date +"%d-%m-%Y_%H:%M:%S")
# echo "[$date_string] starting measurement no. 5 0% tcp"
# do_one_measurement "test-csv/n1-1h-anon.csv" "stats-output/0_tcp_host_${timestamp_str}_5.json" "0"
# date_string=$(date +"%d-%m-%Y_%H:%M:%S")
# echo "[$date_string] measurement no. 5 0% tcp done"


# sleep 60

# date_string=$(date +"%d-%m-%Y_%H:%M:%S")
# echo "[$date_string] starting measurement no. 6 0% tcp"
# do_one_measurement "test-csv/n1-1h-anon.csv" "stats-output/0_tcp_host_${timestamp_str}_6.json" "0"
# date_string=$(date +"%d-%m-%Y_%H:%M:%S")
# echo "[$date_string] measurement no. 6 0% tcp done"

# sleep 60

# date_string=$(date +"%d-%m-%Y_%H:%M:%S")
# echo "[$date_string] starting measurement no. 7 0% tcp"
# do_one_measurement "test-csv/n1-1h-anon.csv" "stats-output/0_tcp_host_${timestamp_str}_7.json" "0"
# date_string=$(date +"%d-%m-%Y_%H:%M:%S")
# echo "[$date_string] measurement no. 7 0% tcp done"

# sleep 60

# date_string=$(date +"%d-%m-%Y_%H:%M:%S")
# echo "[$date_string] starting measurement no. 8 0% tcp"
# do_one_measurement "test-csv/n1-1h-anon.csv" "stats-output/0_tcp_host_${timestamp_str}_8.json" "0"
# date_string=$(date +"%d-%m-%Y_%H:%M:%S")
# echo "[$date_string] measurement no. 8 0% tcp done"

# sleep 60

# date_string=$(date +"%d-%m-%Y_%H:%M:%S")
# echo "[$date_string] starting measurement no. 9 0% tcp"
# do_one_measurement "test-csv/n1-1h-anon.csv" "stats-output/0_tcp_host_${timestamp_str}_9.json" "0"
# date_string=$(date +"%d-%m-%Y_%H:%M:%S")
# echo "[$date_string] measurement no. 9 0% tcp done"

# sleep 60

# date_string=$(date +"%d-%m-%Y_%H:%M:%S")
# echo "[$date_string] starting measurement no. 10 0% tcp"
# do_one_measurement "test-csv/n1-1h-anon.csv" "stats-output/0_tcp_host_${timestamp_str}_10.json" "0"
# date_string=$(date +"%d-%m-%Y_%H:%M:%S")
# echo "[$date_string] measurement no. 10 0% tcp done"
