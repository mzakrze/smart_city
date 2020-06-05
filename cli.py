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
import subprocess
import shlex

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

    parser.add_argument('configuration_file', help='Path to simulation configuration file')

    args = parser.parse_args()

    return args.configuration_file

def parse_config(path):
    # try:
    #     with open(path) as f:
    #         content = f.read()
    # except:
    #     print("Error while reading config file")
    #     raise
    #
    # configParser = configparser.RawConfigParser()
    # configParser.read(path)
    #
    # def trimComment(s):
    #     if "#" not in s:
    #         return s
    #     return s[0:s.index("#")].strip()

    config = Config()
    # config.vpm = int(trimComment(configParser.get('simulation', 'vehicles_per_minute')))
    # config.ip = trimComment(configParser.get('simulation', 'intersection_policy'))
    # config.duration = int(trimComment(configParser.get('simulation', 'duration')))
    config.map_type = "junction_x" # trimComment(configParser.get('simulation', 'map.type'))
    config.map_lanes = 2 # int(trimComment(configParser.get('simulation', 'map.lanes')))
    # config.vehicle_weight = int(trimComment(configParser.get('simulation', 'vehicle.weight')))
    # config.vehicle_power = int(trimComment(configParser.get('simulation', 'vehicle.power')))
    # config.vehicle_braking_force = int(trimComment(configParser.get('simulation', 'vehicle.braking_force')))
    # config.vehicle_max_angular_speed = float(trimComment(configParser.get('simulation', 'vehicle.max_angular_speed')))
    # config.vehicle_max_speed_on_conflict_zone = int(trimComment(configParser.get('simulation', 'vehicle.max_speed_on_conflict_zone')))
    # config.dsrc_loss_p = float(trimComment(configParser.get('simulation', 'dsrc.loss_p')))
    # config.dsrc_avg_latency = int(trimComment(configParser.get('simulation', 'dsrc.avg_latency')))
    # config.random_seed = int(trimComment(configParser.get('simulation', 'random_seed')))

    return config, ""


def insert_to_elastic(graph_raw, config_raw):
    url = 'http://localhost:9200/simulation-info/_doc/' + simulation_name
    myobj = {'simulation_name': simulation_name, "graph_raw": graph_raw, "config_raw": config_raw}

    requests.post(url, data=json.dumps(myobj), headers={'Content-Type': 'application/json; charset=UTF-8'})

    time.sleep(1)  # wait for elastic to index data


def delete_old():
    url = 'http://localhost:9200/simulation-info/_doc/' + simulation_name
    r = requests.delete(url)

def run_command(command):
    # print("Building image ...")
    # os.chdir("algorithm")
    # os.system("docker build -t simulation_algorithm . -q")

    # print("Running simulation ...")
    # run_command("docker run --network smart_city_efk simulation_algorithm")
    #
    # print("CLI Finished.")
    # print("Goto: localhost:3000")
    process = subprocess.Popen(shlex.split(command), stdout=subprocess.PIPE)
    while True:
        if process.poll() is not None:
            break
        output = process.stdout.readline()
        print(output.strip().decode('UTF-8'))
    rc = process.poll()
    return rc

if __name__ == "__main__":
    config_path = read_args()
    print("Reading configuration ...")
    delete_old()
    config, config_raw = parse_config(config_path)

    print("Generating map ...")
    map_raw_json = generate_map_json("junction_x", config.map_lanes)
    insert_to_elastic(map_raw_json, config_raw)

    print("Done.")


