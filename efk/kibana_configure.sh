#!/usr/bin/env bash

curl -X POST "localhost:5601/api/saved_objects/_import" -H "kbn-xsrf: true" --form file=@images/kibana/kibana_conf.ndjson
