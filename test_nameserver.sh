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
    ./test_synch.sh
    echo "WOW WERE FINISHED HERE"
    kill "$pman_PID"
    wait "$pman_PID"
}

do_one_measurement
