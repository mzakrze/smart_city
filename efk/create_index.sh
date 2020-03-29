#!/bin/bash

curl -X DELETE localhost:9200/simulation-log-1
echo ""
curl -X DELETE localhost:9200/simulation-map-1
echo ""
curl -X DELETE localhost:9200/simulation-trip-1

echo ""
echo ""

curl -X PUT localhost:9200/simulation-log-1 \
    -H "Content-Type: application/json" \
    -d '{
        "settings" : {
            "index" : {
                "number_of_shards" : 1, 
                "number_of_replicas" : 0 
            }
        },
        "mappings": {
            "dynamic": "strict",
            "properties": {
                "vehicle_id": {
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
    }'


echo ""
echo ""


curl -X PUT localhost:9200/simulation-map-1 \
    -H "Content-Type: application/json" \
    -d '{
        "settings" : {
            "index" : {
                "number_of_shards" : 1, 
                "number_of_replicas" : 0 
            }
        },
        "mappings": {
            "dynamic": "strict",
            "properties": {
                "vehicle_id": {
                    "type": "integer"
                },
                "location_array": {
                    "type": "geo_point"
                },
                "start_second": {
                    "type": "date",
                    "format": "epoch_second"
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
    }'


echo ""
echo ""

curl -X PUT localhost:9200/simulation-trip-1 \
    -H "Content-Type: application/json" \
    -d '{
        "settings" : {
            "index" : {
                "number_of_shards" : 1, 
                "number_of_replicas" : 0 
            }
        },
        "mappings": {
            "dynamic": "strict",
            "properties": {
                "vehicle_id": {
                    "type": "integer"
                },
                "start_ts": {
                    "type": "date",
                    "format": "epoch_millis"
                },
                "end_ts": {
                    "type": "date",
                    "format": "epoch_millis"
                },
                "origin_lat": {
                    "type": "float"
                },
                "origin_lon": {
                    "type": "float"
                },
                "destination_lat": {
                    "type": "float"
                },
                "destination_lon": {
                    "type": "float"
                } 
            }
        }
    }'


echo ""
echo ""



