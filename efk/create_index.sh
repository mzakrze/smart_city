#!/bin/bash

curl -X DELETE localhost:9200/simulation-1
echo ""

curl -X PUT localhost:9200/simulation-1 \
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
                },
                "my_epoch": {
                    "type": "long"
                }
            }
        }
    }'
echo ""

# test
# curl -X POST localhost:9200/simulation/_doc \
#     -H "Content-Type: application/json" \
#     -d '{"car_id": 5, "location": {"lat": "52.2297700", "lon": "21.0117800"}, "speed": 3.14}'

# curl -X POST localhost:9880/car.xd -d '{"car_no": 5, "location": {"lat": "52.2297700", "lon": "21.0117800"}, "speed": 3.14}'

curl -X PUT localhost:9200/simulation-1-map \
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
                "car_id": {
                    "type": "integer"
                },
                "startSecond:" {
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
                "location": {
                    "type": "geo_point"
                }
            }
        }
    }'
echo ""









