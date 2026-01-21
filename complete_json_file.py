import sys
import re
import json


input_filename = sys.argv[1]
output_filename = sys.argv[2]

with open(output_filename, 'r') as file:
    stat_data = json.load(file)


total_data = {
    "containerOutput" : stat_data,
    "network" : "host"
}
 
with open(output_filename, 'w') as file:
    json.dump(total_data, file, indent=4)
