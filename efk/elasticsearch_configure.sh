#!/bin/bash

curl -X PUT "localhost:9200/_template/template_simulation_map?pretty" -H 'Content-Type: application/json' -d'
{
    "index_patterns": [ "simulation-1-map" ],
    "settings": {
        "number_of_shards": 1,
        "number_of_replicas": 0
    },
    "mappings": {
        "dynamic": "strict",
        "properties": {
            "car_id": {
                "type": "integer"
            },
            "startSecond": {
                "type": "date",
                "format": "epoch_second"
            },
            "location": {
                "type": "geo_point"
            },
            "bbox_north": {
                "type": "float"
            },
            "bbox_south": {
                "type": "float"
            },
            "bbox_east": {
                "type": "float"
            },
            "bbox_west": {
                "type": "float"
            }
        }
    }
}
'

echo ""
echo ""


curl -X PUT "localhost:9200/_template/template_simulation_log?pretty" -H 'Content-Type: application/json' -d'
{
    "index_patterns": [ "simulation-1-log" ],
    "settings": {
        "number_of_shards": 1,
        "number_of_replicas": 0
    },
    "mappings": {
        "dynamic": "strict",
        "properties": {
            "car_id": {
                "type": "integer"
            },
            "location": {
                "type": "geo_point"
            },
            "speed": {
                "type": "float"
            },
            "@timestamp": {
                "type": "date",
                "format": "epoch_millis"
            }
        }
    }
}
'
