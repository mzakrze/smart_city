#!/usr/bin/python3.6

import sys
sys.path.insert(1, 'mapGenerator')
from generate import generate_map_json
import argparse
import configparser
import requests
import json
import time
import os

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
    parser = argparse.ArgumentParser(description='Runs simulation')

    parser.add_argument('--config', help='Path to simulation configuration file', required=True)

    args = parser.parse_args()

    return args.config

def parse_config(path):
    try:
        with open(path) as f:
            content = f.read()
    except:
        print("Error while reading config file")
        raise

    configParser = configparser.RawConfigParser()
    configParser.read(path)

    config = Config()
    config.vpm = int(configParser.get('simulation', 'vehicles_per_minute'))
    config.ip = configParser.get('simulation', 'intersection_policy')
    config.duration = int(configParser.get('simulation', 'duration'))
    config.map_type = configParser.get('simulation', 'map.type')
    config.map_lanes = int(configParser.get('simulation', 'map.lanes'))
    config.vehicle_speed = float(configParser.get('simulation', 'vehicle.max_speed'))
    config.vehicle_acc = float(configParser.get('simulation', 'vehicle.max_acc'))
    config.vehicle_decel = float(configParser.get('simulation', 'vehicle.max_decel'))

    return config, content


def insert_to_elastic(graph_raw, config_raw):
    url = 'http://localhost:9200/simulation-info/_doc/' + simulation_name
    myobj = {'simulation_name': simulation_name, "graph_raw": graph_raw, "config_raw": config_raw}

    requests.post(url, data=json.dumps(myobj), headers={'Content-Type': 'application/json; charset=UTF-8'})

    time.sleep(1)  # wait for elastic to index data


def delete_old():
    url = 'http://localhost:9200/simulation-info/_doc/' + simulation_name
    r = requests.delete(url)

if __name__ == "__main__":
    print("Reading configuration ...")
    config_path = read_args()
    delete_old()
    config, config_raw = parse_config(config_path)

    print("Generating map ...")
    map_raw_json = generate_map_json("junction_x", config.map_lanes)
    insert_to_elastic(map_raw_json, config_raw)

    print("Building image ...")
    os.chdir("algorithm")
    os.system("docker build -t simulation_algorithm . -q")

    print("Running simulation ...")
    os.system("docker run --network smart_city_efk simulation_algorithm")

    print("Finished.")


