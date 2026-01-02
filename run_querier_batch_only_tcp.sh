#! /bin/bash

kill_jobs() {
    for job in $(jobs -p); do
            kill -s SIGTERM $job > /dev/null 2>&1 || (sleep 10 && kill -9 $job > /dev/null 2>&1 &)

    done
}

trap kill_jobs EXIT

do_one_measurement () {
    # The function gets three arguments: input_filename, output_filename, percentage_tcp
    # In the other script the nameserver is started HERE
    date_string=$(date +"%d-%m-%Y_%H:%M:%S")
    echo "[$date_string] Nameserver should be starting now"
    # The while loop should sync the two processes every time, so they don't get out of loop too much
    while [ ! -f "/tmp/test-readiness/ready" ]; do
        sleep 1 # Every second check whether it was created yet or not
    done
    date_string=$(date +"%d-%m-%Y_%H:%M:%S")
    echo "[$date_string] Nameserver has signalled it has started"
    rm -f "/tmp/test-readiness/ready" # remove the file so it's not there on the next run
    # In the other server the statistics is started HERE
    # date_string=$(date +"%d-%m-%Y_%H:%M:%S")
    # echo "[$date_string] Statistics should have been started"
    # # Wait for the statistics of the other server to finish
    # sleep 60
    date_string=$(date +"%d-%m-%Y_%H:%M:%S")
    echo "[$date_string] Starting querier"
    ./no-retry/querier_no_retry -f "$1" -s "$3" > "$4"
    date_string=$(date +"%d-%m-%Y_%H:%M:%S")
    echo "[$date_string]Querier done, touching the file on eline02"
    ssh eline@145.102.0.142 "touch /tmp/test-readiness/ready"
    # Once the querier has finished we gather statistics for 60 seconds again
    sleep 60
}

timestamp_str=$(date +"%d-%m-%Y_%H:%M:%S")

measurements_to_do=(1 2 3 4 5 6 7 8 9 10)

for measurement in "${measurements_to_do[@]}"
do 
    date_string=$(date +"%d-%m-%Y_%H:%M:%S")
    echo "[$date_string] starting measurement no. $measurement nsd"
    do_one_measurement "test-csv/n1-1h-anon.csv" "100" "145.102.0.142:4242" "output/no-retries/querier_nsd_${timestamp_str}_$measurement.out"
    date_string=$(date +"%d-%m-%Y_%H:%M:%S")
    echo "[$date_string] measurement no. $measurement nsd done"
    sleep 60
done

for measurement in "${measurements_to_do[@]}"
do 
    date_string=$(date +"%d-%m-%Y_%H:%M:%S")
    echo "[$date_string] starting measurement no. $measurement knot"
    do_one_measurement "test-csv/n1-1h-anon.csv" "100" "145.102.0.142:4242" "output/no-retries/querier_knot_${timestamp_str}_$measurement.out"
    date_string=$(date +"%d-%m-%Y_%H:%M:%S")
    echo "[$date_string] measurement no. $measurement knot done"
    sleep 60
done

for measurement in "${measurements_to_do[@]}"
do 
    date_string=$(date +"%d-%m-%Y_%H:%M:%S")
    echo "[$date_string] starting measurement no. $measurement bind"
    do_one_measurement "test-csv/n1-1h-anon.csv" "100" "145.102.0.142:4242" "output/no-retries/querier_bind_${timestamp_str}_$measurement.out"
    date_string=$(date +"%d-%m-%Y_%H:%M:%S")
    echo "[$date_string] measurement no. $measurement bind done"
    sleep 60
done

for measurement in "${measurements_to_do[@]}"
do 
    date_string=$(date +"%d-%m-%Y_%H:%M:%S")
    echo "[$date_string] starting measurement no. $measurement pdns"
    do_one_measurement "test-csv/n1-1h-anon.csv" "100" "145.102.0.142:4242" "output/no-retries/querier_pdns_${timestamp_str}_$measurement.out"
    date_string=$(date +"%d-%m-%Y_%H:%M:%S")
    echo "[$date_string] measurement no. $measurement pdns done"
    sleep 60
done