import sys
import re
import json

def parse_querier_output(querier_output):
    # Find all lines of the form "something: value"
    number_errors = 0
    pairs = [line.split(":", 1) for line in querier_output if ":" in line]

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
        elif key.lower().startswith("sendqueries"):
            number_errors += 1
        else:
            # Parse regular integers/floats
            num_match = re.search(r'[\d.]+', value)
            results[key] = float(num_match.group()) if num_match else None
    results["number_errors"] = number_errors
    return results


input_filename = sys.argv[1]
output_filename = sys.argv[2]
# querier_output_file = sys.argv[3]


#with open(querier_output_file, 'r') as file:
#    querier_output = file.readlines() # This will put each line in the file as a separate string in an array

#parsed_output = parse_querier_output(querier_output)

with open(output_filename, 'r') as file:
    stat_data = json.load(file)


total_data = {
    "containerOutput" : stat_data,
 #   "numberErrors" : parsed_output['number_errors'],
 #   "querySendingTime" : parsed_output['Execution time'],
 #   "numberOfQueries" : parsed_output['Total queries'],
 #   "tcpPercentage" : parsed_output['Amount with TCP']/parsed_output['Total queries'] * 100,
 #   "rawOutput" : querier_output,
    "network" : "host"
}

with open(output_filename, 'w') as file:
    json.dump(total_data, file, indent=4)
