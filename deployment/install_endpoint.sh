#!/bin/bash

set -e # fail if unexpected error


echo "Started installation of 2SMS Endpoint application"

# Defaults
INSTALLATION_PATH="$HOME/2SMS/deployment/endpoint"
DEFAULT_SOCKET=/run/shm/sciond/default.sock
SERVICE_FILE_NAME=2SMSendpoint.service
SERVICE_FILE_LOCATION=/etc/systemd/system
MANAGER_IP='192.33.93.196'
IP=$(curl -s ipinfo.io/ip)
[ -f $SC/gen/ia ] && IA=$(cat $SC/gen/ia | sed 's/_/:/g') || { echo "Missing $SC/gen/ia file"; exit 1; }

# Prepare filesystem
echo "Creating installation directory at $INSTALLATION_PATH"
mkdir -p $INSTALLATION_PATH
cd $INSTALLATION_PATH

# Download binary file
echo "Downloading binary"
rm -f endpoint
git clone https://gist.github.com/baehless/0af8c4fca2db16737a6e31b7e725ad98
mv 0af8c4fca2db16737a6e31b7e725ad98/endpoint .
rm -rf 0af8c4fca2db16737a6e31b7e725ad98

# Make executable
chmod +x endpoint

wget https://gist.github.com/juagargi/376323076d37bf319ec29eb2b0a071f4/raw/3fb9efbb630d8d28b9f880d33024e5b0eb05e004/endpoint-deployment.tgz -O endpoint-deployment.tgz
tar xf endpoint-deployment.tgz
if [ ! -f ca_certs/bootstrap.json ] || [ ! -f ca_certs/ca.crt ] || [ ! -f ca_certs/ISD*AS*.crt ] || [ ! -f auth/model.conf ]; then
    echo "ca_certs/ or auth/ files missing after unpacking endpoint-deployment.tgz from our gist"
    exit 1
fi

# Download service file
echo "Downloading service file"
rm -f $SERVICE_FILE_NAME
wget https://raw.githubusercontent.com/netsec-ethz/2SMS/master/endpoint/2SMSendpoint.service -O $SERVICE_FILE_NAME

# Modify service file with correct SCION address and IP parameters
sed -i -r "s/_USER_/$USER/g;s/_MANAGER_IP_/$MANAGER_IP/g;s/^(.+)_IA_,\[_IP_\]:9199 -endpoint\.IP _IP_(.+)$/\1$IA,[$IP]:9199 -endpoint.IP $IP\2/g" $SERVICE_FILE_NAME

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
