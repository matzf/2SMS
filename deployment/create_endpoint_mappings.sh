#!/bin/bash

set -e

if [ $# -gt 1 ]; then
    echo "Expecting at most 1 argument but received $#"
    echo "$0 services_dir"
    exit 1
elif [ $# -eq 1 ]; then
    services_dir="$1"
else
    services_dir=$(find $SC/gen -type d -regex ".*/ISD.*/AS.*" | head -n 1)
fi

services=$(ls $services_dir)
mapping_file='mappings.json'

[ -f $mapping_file ] && { echo "$mapping_file already exists, skipping initialization"; exit 0; }

touch $mapping_file

# Find mappings
echo "Looking for mappings in '$services_dir'"
declare -a mappings
mappings+='{"Path":"/node","Port":"9100"}'
for s in $services; do
    service_dir="$services_dir/$s"
    if [ -d $service_dir ] && { [[ $s =~ ^br[0-9]+.*$ ]] || [[ $s =~ ^ps[0-9]+.*$ ]] || [[ $s =~ ^bs[0-9]+.*$ ]] || [[ $s =~ ^cs[0-9]+.*$ ]]; }; then
        # TODO: changes to cs and ps parsing (will be in a .toml file instead of the supervisord.conf)
        # Don't add cs and ps instances until they expose metrics
        if [[ $s =~ ^cs.*$ ]]; then
            echo "Skipping certificate server ('$service_dir') because it doesn't yet expose metrics."
            continue
        elif [[ $s =~ ^ps.*$ ]]; then
            echo "Skipping path server ('$service_dir') because it doesn't yet expose metrics."
            continue
        fi
        supervisor_file="$service_dir/supervisord.conf"
        if [ -f $supervisor_file ]; then
            # Find port from `prom` argument in supervisor.conf file
            port=$(sed -ne '/.*-prom\>[^:]*:\([0-9]*\).*/{s//\1/p;q}' $supervisor_file)
            name="$(sed -r 's/^([a-z]+)[0-9]+-.*-([0-9]+)$/\1-\2/' <<< $s)"
            if [ ! -z $port ]; then
                mapping='{"Path":"/'$name'","Port":"'$port'"}'
                mappings+=($mapping)
                echo "Added $mapping"
            else
                echo "No 'prom' argument found in '$supervisor_file'"
            fi
        else
            echo "'$supervisor_file' not found"
        fi
    else
        echo "'$service_dir' is not of interest for creating mappings"
    fi
done
map_len=${#mappings[@]}

# Write to file
echo "Writing mappings to file"
echo "[" >> $mapping_file
for ((i=0; i < $((map_len - 1)); i++)); do
    echo "  ${mappings[$i]}," >> $mapping_file
done
echo "  ${mappings[$(($map_len - 1))]}" >> $mapping_file
echo "]" >> $mapping_file

echo "Finished writing to 'mappings.json'"