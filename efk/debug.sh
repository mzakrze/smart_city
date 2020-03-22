#!/bin/bash


# curl localhost:9200/simulation/_search?pretty \
# -H "Content-Type: application/json" \
# -d '
# {
#     "aggs" : {
#        "min_date": {"min": {"field": "@timestamp"}},
#        "max_date": {"max": {"field": "@timestamp"}}
#     }
# }'




curl localhost:9200/simulation/_search?pretty \
-H "Content-Type: application/json" \
-d '
{
    "query": {
        "bool": {
            "filter": {
                "range": {
                    "@timestamp": {"gte": 0, "lte": 991584882223417 }
                }
            }
        }
    },
    "sort": [
        { "@timestamp": {"order": "asc" }},
        { "car_id": {"order": "asc" }}
    ]
}'