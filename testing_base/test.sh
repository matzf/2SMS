#!/bin/bash

BASEDIR=$(dirname $(realpath $0))

fail_test() {
    MSG="$1"
    echo "$MSG"
    echo "Test failed."
    exit 1
}


cd "$BASEDIR"

curl http://localhost:9999/targets -X POST -H 'Content-Type: application/json' -d '{"Name":"deleteme1","ISD":"1","AS":"ffaa:1:a","IP":"127.0.0.1","Port":"32041","Path":"/bs","Labels":{}}]}' -s >/dev/null
grep deleteme1 ./scraper/prometheus/prometheus.yml >/dev/null || fail_test "Cannot find 'deleteme1' in the prometheus.yml config file"


echo "tests ok."
