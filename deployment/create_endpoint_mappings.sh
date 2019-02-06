#!/bin/bash

set -e

if [ $# -gt 1 ]; then
    echo "Expecting at most 1 argument but received $#"
    echo "$0 services_dir"
    exit 1
elif [ $# -eq 1 ]; then
    services_dir="$1"
else
    services_dir=$(find $SC/gen/ -type d -regex ".*/ISD.*/AS.*" | head -n 1)
fi

services=$(ls $services_dir)
mapping_file='mappings.json'

[ -f $mapping_file ] && { echo "$mapping_file already exists, skipping initialization"; exit 0; }

# Find mappings
echo "Looking for mappings in '$services_dir'"
declare -a mappings
# Add default mappings
mappings+='{"Path":"/node","Port":"9100"}'

for s in $services; do
    service_dir="$services_dir/$s"
    if [ -d $service_dir ]; then
        port=""
        name=""
        config_file=""
        is_toml=true
        # Check if the directory corresponds to a service exposing metrics
        if [[ $s =~ ^cs[0-9]+.*$ ]] || [[ $s =~ ^ps[0-9]+.*$ ]] || [[ $s = endhost ]]; then
            if [[ $s = endhost ]]; then
                name="sciond"
                config_file="$service_dir/sciond.toml"
            else
                name="$(sed -r 's/^([a-z]+)[0-9]+-.*-([0-9]+)$/\1-\2/' <<< $s)"
                config_file="$service_dir/${name:0:2}config.toml"
            fi
        elif [[ $s =~ ^br[0-9]+.*$ ]] || [[ $s =~ ^bs[0-9]+.*$ ]]; then
            is_toml=false
            name="$(sed -r 's/^([a-z]+)[0-9]+-.*-([0-9]+)$/\1-\2/' <<< $s)"
            config_file="$service_dir/supervisord.conf"
        else
            echo "'$service_dir' is not of interest for creating mappings"
            continue
        fi

        # Parse the configuration file to find the port where the metrics are exposed
        echo "Looking for $config_file"
        if [ -f $config_file ]; then
            if $is_toml; then
                # Find port from `Prometheus` key in the service .toml config file
                port=$(sed -ne '/^Prometheus *= *".*:\([0-9]*\)".*$/{s//\1/p;q}'  $config_file)
            else
                # Find port from `prom` argument in the service supervisord.conf file
                port=$(sed -ne '/.*-prom\>[^:]*:\([0-9]*\).*/{s//\1/p;q}' $config_file)
            fi
        else
            echo "'$config_file' not found"
            continue
        fi

        # Create the mapping corresponding to the port
        if [ ! -z $port ]; then
            mapping='{"Path":"/'$name'","Port":"'$port'"}'
            mappings+=($mapping)
            echo "Added $mapping"
        else
            echo "No prometheus-related parameter found in '$config_file'"
        fi
    fi
done
map_len=${#mappings[@]}

# Write all mappings to file
echo "Writing mappings to file"
touch $mapping_file
echo "[" >> $mapping_file
for ((i=0; i < $((map_len - 1)); i++)); do
    echo "  ${mappings[$i]}," >> $mapping_file
done
echo "  ${mappings[$(($map_len - 1))]}" >> $mapping_file
echo "]" >> $mapping_file

echo "Finished writing to 'mappings.json'"