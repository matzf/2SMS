#!/bin/bash

set -e # fail if unexpected error


echo "Started installation of 2SMS Endpoint application"

# Defaults
INSTALLATION_PATH="$HOME/2SMS/deployment/endpoint"
DEFAULT_SOCKET=/run/shm/sciond/default.sock
SERVICE_FILE_NAME=2SMSendpoint.service
SERVICE_FILE_LOCATION=/etc/systemd/system
MANAGER_IP='129.132.85.121'
monitoring_download_page="monitoring.scionlab.org/downloads/public/endpoint"
IP=$(curl -s ipinfo.io/ip)
[ -f $SC/gen/ia ] && IA=$(cat $SC/gen/ia | sed 's/_/:/g') || { echo "Missing $SC/gen/ia file"; exit 1; }

# Prepare filesystem
echo "Creating installation directory at $INSTALLATION_PATH"
mkdir -p $INSTALLATION_PATH
cd $INSTALLATION_PATH

# Download latest endpoint binary
echo "Downloading endpoint binary"
rm -f endpoint
wget "https://$monitoring_download_page/endpoint.tar.gz" >/dev/null 2>&1
tar -xzvf endpoint.tar.gz
rm -f endpoint.tar.gz
chmod +x endpoint

# Download node exporter binary
echo "Downloading node exporter binary"
rm -rf node-exporter/node_exporter
wget "https://$monitoring_download_page/node_exporter.tar.gz" >/dev/null 2>&1
tar -xzvf node_exporter.tar.gz
rm -f node_exporter.tar.gz
chmod +x node_exporter
mkdir -p node-exporter
mv node_exporter node-exporter/

# Download configuration files
echo "Downloading configuration files"
wget "https://$monitoring_download_page/endpoint_configuration.tar.gz" >/dev/null 2>&1
tar -xzvf endpoint_configuration.tar.gz
rm -f endpoint_configuration.tar.gz
mv -n configuration/* .
rm -rf configuration

# Check contents
if [ ! -f ca_certs/bootstrap.json ] || [ ! -f ca_certs/ca.crt ] || [ ! -f ca_certs/ISD*AS*.crt ] || [ ! -f auth/model.conf ] || [ ! -f 2SMSendpoint.service ] || [ ! -f create_endpoint_mappings.sh ]; then
    echo "Some required configuration file is missing from the unpacked endpoint_configuration archive. Please make sure that you have the following files: ca_certs/boostrap.json, ca_certs/ca.crt, ca_certs/ISD*AS*.crt, auth/model.conf, 2SMSendpoint.service, create_endpoint_mappings.sh"
    exit 1
fi
chmod +x create_endpoint_mappings.sh

# Modify service file with correct SCION address and IP parameters
sed -i -r "s/_USER_/$USER/g;s/_MANAGER_IP_/$MANAGER_IP/g;s/^(.+)_IA_,\[_IP_\]:9199 (.+)$/\1$IA,[$IP]:9199 \2/g" $SERVICE_FILE_NAME

# Move service file to the right place
sudo mv $SERVICE_FILE_NAME /etc/systemd/system/
sudo chown root "$SERVICE_FILE_LOCATION/$SERVICE_FILE_NAME"
sudo chgrp root "$SERVICE_FILE_LOCATION/$SERVICE_FILE_NAME"

# Check system
echo "Checking SCION installation"
if [ -z $SC ]; then
    echo 'SCION environment variable $SC not set, is SCION installed correctly?'
    exit 1
elif [ ! -d $SC/gen/ISD*/AS*/endhost/certs ]; then
    echo "SCION endhost's cert directory not found. Aborting."
    exit 1
elif [ ! -S $DEFAULT_SOCKET ]; then
    echo "SCION default socket not found at $DEFAULT_SOCKET"
    exit 1
else
    echo "SCION correctly installed"
fi

# Create mappings file
echo "Removing previous mappings file"
rm -f mappings.json
./create_endpoint_mappings.sh

# Start service
echo "Stopping $SERVICE_FILE_NAME"
sudo systemctl stop $SERVICE_FILE_NAME || true
echo "Reloading Daemons"
sudo systemctl daemon-reload
echo "Starting $SERVICE_FILE_NAME"
sudo systemctl start $SERVICE_FILE_NAME
echo "Enabling $SERVICE_FILE_NAME at boot time"
sudo systemctl enable $SERVICE_FILE_NAME || true # at boot time

# Check service status
if systemctl is-active --quiet 2SMSendpoint.service; then
    echo "Successfully installed and started 2SMS Endpoint application"
else
    echo 'Failed starting 2SMS Endpoint application, something went wrong during installation. Check journal and syslog to have more details.'
    exit 1
fi
