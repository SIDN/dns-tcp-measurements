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
    sleep 240 # wait a bit for the zonefile to load
    date_string=$(date +"%d-%m-%Y_%H:%M:%S")
    echo "[$date_string] Nameserver should have been started"
    # In the other server the statistics is started HERE
    date_string=$(date +"%d-%m-%Y_%H:%M:%S")
    echo "[$date_string] Statistics should have been started"
    # Wait for the statistics of the other server to finish
    sleep 10
    date_string=$(date +"%d-%m-%Y_%H:%M:%S")
    echo "[$date_string] Starting querier"
    ~/go/bin/querier -f "$1" -p "$2" -s "$3" > "$4"
    date_string=$(date +"%d-%m-%Y_%H:%M:%S")
    echo "[$date_string]Querier done"
    # Once the querier has finished we gather statistics for 60 seconds again
    sleep 10
}


timestamp_str=$(date +"%d-%m-%Y_%H:%M:%S")

do_one_measurement "test-csv/n1-1h-anon.csv" "100" "145.102.0.142:4242" "output/querier_100_tcp_${timestamp_str}.out"
date_string=$(date +"%d-%m-%Y_%H:%M:%S")
echo "[$date_string] measurement 100% tcp done"
# sleep 10
# do_one_measurement "/home/elmer/ns1data/ns1-1h-anon.csv" "stats-output/100_tcp_host_${timestamp_str}.json" "100"
# echo "measurement 100% tcp done"

