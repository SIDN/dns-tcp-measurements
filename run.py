import subprocess
import time
import re
import json
from datetime import datetime

def parse_querier_output(querier_output):
    # Find all lines of the form "something: value"
    pairs = [line.split(":", 1) for line in querier_output.splitlines() if ":" in line]

    results = {}
    for key, value in pairs:
        key = key.strip()
        value = value.strip()

        # Special handling for the execution time line
        if key.lower().startswith("execution time"):
            # Handle both "XXms" and "XmYY.s" formats
            match = re.match(r'(?:(\d+)m)?([\d.]+)(ms|s)?', value)
            if match:
                minutes = int(match.group(1)) if match.group(1) else 0
                seconds = float(match.group(2))
                unit = match.group(3)

                # Normalize everything to seconds
                if unit == "ms":
                    total_seconds = seconds / 1000
                else:
                    total_seconds = minutes * 60 + seconds

                results[key] = total_seconds
            else:
                results[key] = None
        else:
            # Parse regular integers/floats
            num_match = re.search(r'[\d.]+', value)
            results[key] = float(num_match.group()) if num_match else None
    return results

def do_one_measurement(input_filename, output_filename, percentage_tcp):
    # Start the Podman container (non-blocking)
    podman_process = subprocess.Popen([
        "podman", "run", "--replace", "--rm", "--network=host", 
        "-v", "./nsd/nsd.conf:/etc/nsd/nsd.conf",
        "-v", "./nsd/zones:/dns",
        "-v", "./nsd/config:/config",
        "--name", "nsd-query", "querier-nsd:latest",
        "nsd", "-V", "2", "-d"
    ])
    # Give the container a moment to start up
    time.sleep(2)
    # Start the stats.sh process
    stats_process = subprocess.Popen([
            "./stats.sh", output_filename
        ],
        stdout=subprocess.DEVNULL,
        stderr=subprocess.STDOUT
    )
    # Gather container information for 60 seconds
    time.sleep(60)
    # Query the NSD server using the querier script
    result = subprocess.run(
        "~/go/bin/querier -f " + input_filename + " -p " + percentage_tcp,
        shell=True, 
        capture_output=True, text=True
    )
    # Wait an addition 60 seconds to gather more container information
    time.sleep(60)
    # Terminates the stats.sh script that also finishes writing output to the output file then
    stats_process.terminate()

    # Terminate the nsd server
    podman_process.terminate()
    print("Querier output:")
    print(result.stdout)
    print(result.stderr)

    parsed_output = parse_querier_output(result.stdout)

    with open(output_filename, 'r') as file:
        stat_data = json.load(file)

    total_data = {
        "containerOutput" : stat_data,
        "querySendingTime" : parsed_output['Execution time'],
        "numberOfQueries" : parsed_output['Total queries'],
        "tcpPercentage" : parsed_output['Amount with TCP']/parsed_output['Total queries'] * 100,
        "network" : "host"
    }

    with open(output_filename, 'w') as file:
        json.dump(total_data, file, indent=4)

timestamp_str = datetime.now().strftime("%d-%m-%Y_%H:%M:%S")
# We first do one measurement over UDP
do_one_measurement("/home/elmer/ns1data/ns1-1h-anon.csv", "stats-output/tcp_host_" + timestamp_str + ".json", "100")
do_one_measurement("/home/elmer/ns1data/ns1-1h-anon.csv",  "stats-output/udp_host_" + timestamp_str + ".json", "0")
