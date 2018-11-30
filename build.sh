#!/bin/bash

set -e

BASEDIR=$(dirname $(realpath $0))
cd "$BASEDIR"

echo "Building manager"
cd manager
go build
echo "Building scraper"
cd ../scraper
go build
echo "Building endpoint"
cd ../endpoint
go build
