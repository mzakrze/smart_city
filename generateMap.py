#!/usr/bin/python3.6

import sys
sys.path.insert(1, 'mapGenerator')
from generate import generate_map_json
import argparse
import requests
import json
import time

simulation_name = "simulation"

class Config:
    def __init__(self):
        self.config = None
        self.vpm = None
        self.ip = None
        self.duration = None
        self.map_type = None
        self.map_lanes = None
        self.vehicle_speed = None
        self.vehicle_acc = None
        self.vehicle_decel = None


def read_args():
    parser = argparse.ArgumentParser(description='Generates map and inserts to Elasticsearch')

    parser.add_argument('--lanes', help='Number of lanes in the intersection', type=int, required=True)

    args = parser.parse_args()

    return args.lanes


def insert_to_elastic(graph_raw):
    url = 'http://localhost:9200/simulation-info/_doc/' + simulation_name
    myobj = {'simulation_name': simulation_name, "graph_raw": graph_raw}

    try:
        requests.post(url, data=json.dumps(myobj), headers={'Content-Type': 'application/json; charset=UTF-8'})
    except requests.ConnectionError:
        print("Error while trying to connect to Elasticsearch (localhost:9200)")
        print("Make sure ES is up and running")
        sys.exit(1)

    time.sleep(1)  # wait for elastic to index data


def delete_old():
    url = 'http://localhost:9200/simulation-info/_doc/' + simulation_name
    try:
        r = requests.delete(url)
    except requests.ConnectionError:
        print("Error while trying to connect to Elasticsearch (localhost:9200)")
        print("Make sure ES is up and running")
        sys.exit(1)

if __name__ == "__main__":
    map_lanes = read_args()
    print("Deleting old map from Elasticsearch ...")
    delete_old()

    print("Generating map ...")
    map_raw_json = generate_map_json("junction_x", map_lanes)
    print("Indexing generated map in Elasticsearch ...")
    insert_to_elastic(map_raw_json)

    print("Done.")


