#!/bin/bash

sleep 20 # wait for elastic

curl -X DELETE elastic:9200/simulation-map
echo ""
curl -X DELETE elastic:9200/simulation-vehicle
echo ""

curl -X PUT "elastic:9200/simulation-map" -H 'Content-Type: application/json' -d'
{
    "settings": {
        "number_of_shards": 1,
        "number_of_replicas": 0
    },
    "mappings": {
        "dynamic": "strict",
        "properties": {
            "simulation_name": {
                "type": "keyword"
            },
            "vehicle_id": {
                "type": "integer"
            },
            "second": {
                "type": "date",
                "format": "epoch_second"
            },
            "location_array": {
                "type": "geo_point"
            },
            "alpha_array": {
                "type": "float"
            },
            "state_array": {
                "type": "float"
            }
        }
    }
}
'

echo
echo
echo
echo


curl -X PUT "elastic:9200/simulation-vehicle" -H 'Content-Type: application/json' -d'
{
    "settings": {
        "number_of_shards": 1,
        "number_of_replicas": 0
    },
    "mappings": {
        "dynamic": "strict",
        "properties": {
            "simulation_name": {
                "type": "keyword"
            },
            "vehicle_id": {
                "type": "integer"
            },
            "way_from": {
                "type": "integer"
            },
            "way_to": {
                "type": "integer"
            },
            "start_time": {
                "type": "date",
                "format": "epoch_second"
            },
            "finish_time": {
                "type": "date",
                "format": "epoch_second"
            },
            "duration": {
                "type": "integer"
            },
            "speed_array": {
                "type": "float"
            },
            "acc_array": {
                "type": "float"
            }
        }
    }
}
'


echo
echo
echo
echo





