#!/bin/bash

BASEDIR=$(dirname $(realpath $0))
cd "$BASEDIR/.."

# defaults:
[ -f $SC/gen/ia ] && IA=$(cat $SC/gen/ia | sed 's/_/:/g')
[ -x ./manager/manager ] && MANAGER=./manager/manager
[ -x ./scraper/scraper ] && SCRAPER=./scraper/scraper
[ -x ./endpoint/endpoint ] && ENDPOINT=./endpoint/endpoint

if [ ! $# -eq 0 ] && [ ! $# -eq 4 ]; then
    echo "Expecting 4 arguments but received $#"
    echo "$0 IA manager scraper endpoint"
    exit 1 
elif [ $# -eq 4 ]; then
    IA=$1
    MANAGER=$2
    SCRAPER=$3
    ENDPOINT=$4
fi

if [ -z "$MANAGER" ] || [ -z "$SCRAPER" ] || [ -z "$ENDPOINT" ]; then
    echo "Missing at least one executable. Invoke the script with parameters!"
    exit 1
fi

# Set variables
IP=127.0.0.1

[ -x ./testing_base/stop.sh ] && { echo 'Stopping previous runs'; ./testing_base/stop.sh; }

echo 'Copying executables'
cp $MANAGER $BASEDIR/manager/manager
# scraper
cp $SCRAPER $BASEDIR/scraper/scraper
mkdir -p $BASEDIR/scraper/prometheus
cp $(dirname $SCRAPER)/prometheus/prometheus $BASEDIR/scraper/prometheus/
touch $BASEDIR/scraper/prometheus/prometheus.yml
# endpoint
cp $ENDPOINT $BASEDIR/endpoint/endpoint
mkdir -p $BASEDIR/endpoint/auth
cp $(dirname $ENDPOINT)/auth/model.conf $BASEDIR/endpoint/auth
mkdir -p $BASEDIR/endpoint/node-exporter
cp $(dirname $ENDPOINT)/node-exporter/node_exporter $BASEDIR/endpoint/node-exporter/

echo 'Starting manager application'
cd $BASEDIR/manager
rm -r scrapers.json
rm -r endpoints.json
rm -r storages.json
rm -f approved_certs/*
rm -f ca/*
rm -f auth/manager.*
./manager -local "$IA",[$IP]:0 -manager.IP $IP --ports.management 10002 --ports.no-client-verif 10000 --ports.client-verif 10001 > ../manager.out  2>&1 &
sleep 1
curl http://$IP:10002/manager/signing/enable && echo "" # || exit 1

echo 'Copying certificate and bootstrap files'
rm -f ../scraper/auth/scraper.*
mkdir -p ../scraper/ca_certs
cp ca/ca.crt ../scraper/ca_certs
cp ca/bootstrap.json ../scraper/ca_certs
rm -f ../endpoint/auth/endpoint.*
mkdir -p ../endpoint/ca_certs
cp ca/ca.crt ../endpoint/ca_certs
cp ca/bootstrap.json ../endpoint/ca_certs

echo 'Starting scraper application'
cd ../scraper
./scraper -local "$IA",[$IP]:0 -scraper.IP $IP -scraper.ports.management 9900 -scraper.ports.local 9999 -manager.IP $IP -manager.unverif-port 10000 -manager.verif-port 10001 > ../scraper.out 2>&1 &

echo 'Starting endpoint application'
cd ../endpoint
./endpoint -local "$IA",[$IP]:9199  -endpoint.IP $IP -endpoint.enable-node true -manager.IP $IP  -endpoint.ports.local 9998 -endpoint.ports.management 9905 > ../endpoint.out 2>&1 &
sleep 1
ps 
echo "All started"
