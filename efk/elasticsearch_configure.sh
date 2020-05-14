#!/bin/bash


curl -X PUT "localhost:9200/simulation-map" -H 'Content-Type: application/json' -d'
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
            }
        }
    }
}
'

echo
echo
echo
echo


curl -X PUT "localhost:9200/simulation-vehicle" -H 'Content-Type: application/json' -d'
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


curl -X PUT "localhost:9200/simulation-info" -H 'Content-Type: application/json' -d'
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
            "simulation_started_ts": {
                "type": "date",
                "format": "epoch_second"
            },
            "simulation_finished_ts": {
                "type": "date",
                "format": "epoch_second"
            },
            "simulation_max_ts": {
                "type": "integer"
            },
            "config_raw": {
                "type": "text"
            },
            "map_nodes": {
                "type": "text"
            },
            "map_edges": {
                "type": "text"
            }
        }
    }
}
'

echo
echo
echo
echo


curl -X PUT "localhost:9200/simulation-intersection" -H 'Content-Type: application/json' -d'
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
            "second": {
                "type": "date",
                "format": "epoch_second"
            },
            "arrive_no": {
                "type": "integer"
            },
            "leave_no": {
                "type": "integer"
            }
        }
    }
}
'

echo
echo
echo
echo



curl -X PUT "localhost:9200/simulation-vehiclestep" -H 'Content-Type: application/json' -d'
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
            "ts": {
                "type": "date",
                "format": "epoch_second"
            },
            "speed": {
                "type": "float"
            },
            "acc": {
                "type": "float"
            }
        }
    }
}
'

echo
echo
echo

