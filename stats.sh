#! /bin/bash

# Check if at least one argument is passed
if [ $# -lt 1 ]; then
    echo "Usage: $0 <output_filename>"
    exit 1
fi


OUTPUT_FILE="$1"

echo "[" > $OUTPUT_FILE

first_entry=true

trap "echo ']' >> $OUTPUT_FILE; exit" EXIT

for (( ; ; )) 
do 
    # Get the statistics
    # TODO check whether this command is also correct on the testbed
    response=$(curl --unix-socket /run/user/1010/podman/podman.sock http://d/v5.0.0/libpod/containers/stats?stream=false) 
    if [ "$first_entry" = true ]; then 
        echo "$response" >> $OUTPUT_FILE
        first_entry=false
    else
        echo ", $response" >> $OUTPUT_FILE
    fi

    sleep 30
done 
