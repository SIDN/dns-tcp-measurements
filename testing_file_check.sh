#! /bin/bash

while [ ! -f "/tmp/test-readiness/ready" ]; do
    sleep 1 # Every second check whether it was created yet or not
done

echo "done"
