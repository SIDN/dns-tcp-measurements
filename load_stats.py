import json

# TODO: this needs to be extended to handle 

file_path = 'container_stats.json'


with open(file_path, 'r') as file:
    data = json.load(file)  

print(data)