#!/bin/bash

sleep 20 # wait for elastic

curl -X DELETE elastic:9200/simulation-log-1
echo ""
curl -X DELETE elastic:9200/simulation-map-1
echo ""
curl -X DELETE elastic:9200/simulation-trip-1

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


curl -X PUT "elastic:9200/_template/simulation-vehicle-template" -H 'Content-Type: application/json' -d'
{
    "index_patterns": ["simulation-vehicle"],
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


curl -X PUT "elastic:9200/_template/simulation-info-template" -H 'Content-Type: application/json' -d'
{
    "index_patterns": ["simulation-info"],
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
            "throughput": {
                "type": "integer"
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
            "graph_raw": {
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


curl -X PUT "elastic:9200/_template/simulation-intersection-template" -H 'Content-Type: application/json' -d'
{
    "index_patterns": ["simulation-intersection"],
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



curl -X PUT "elastic:9200/_template/simulation-vehiclestep-template" -H 'Content-Type: application/json' -d'
{
    "index_patterns": ["simulation-vehiclestep"],
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
