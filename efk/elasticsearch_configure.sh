#!/bin/bash

# TODO - dodac to do entry-point.sh, sprawdzac

curl -X PUT "localhost:9200/_template/template_simulation_map?pretty" -H 'Content-Type: application/json' -d'
{
    "index_patterns": [ "simulation-map-1" ],
    "settings": {
        "number_of_shards": 1,
        "number_of_replicas": 0
    },
    "mappings": {
        "dynamic": true,
        "properties": {
            "vehicle_id": {
                "type": "integer"
            },
            "start_second": {
                "type": "date",
                "format": "epoch_second"
            },
            "location_array": {
                "type": "geo_point"
            },
            "alpha_array": {
                "type": "float"
            },
        }
    }
}
'

echo ""
echo ""


curl -X PUT "localhost:9200/_template/template_simulation_log?pretty" -H 'Content-Type: application/json' -d'
{
    "index_patterns": [ "simulation-log-1" ],
    "settings": {
        "number_of_shards": 1,
        "number_of_replicas": 0
    },
    "mappings": {
        "dynamic": true,
        "properties": {
            "vehicle_id": {
                "type": "integer"
            },
            "location": {
                "type": "geo_point"
            },
            "alpha": {
                "type": "float"
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

# FIXME - template for "trip"