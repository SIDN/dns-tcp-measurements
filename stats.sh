#! /bin/bash

OUTPUT_FILE="container_stats.json"

echo "[" > $OUTPUT_FILE

first_entry=true

trap "echo ']' >> $OUTPUT_FILE; exit" SIGINT

# Infinite loop:
for (( ; ; )) 
do 
    # Get the statistics
    # TODO check whether this command is also correct on the testbed
    response=$(curl --unix-socket /run/user/1000/podman/podman.sock http://d/v5.0.0/libpod/containers/stats?stream=false) 
    if [ "$first_entry" = true ]; then 
        echo "$response" >> $OUTPUT_FILE
        first_entry=false
    else
        echo ", $response" >> $OUTPUT_FILE
    fi

    sleep 2
done 
