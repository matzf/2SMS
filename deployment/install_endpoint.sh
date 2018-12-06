#!/bin/bash

echo "Started installation of 2SMS Endpoint application"

INSTALLATION_PATH=/home/scion/2SMS/deployment/endpoint
DEFAULT_SOCKET=/run/shm/sciond/default.sock
SERVICE_FILE=2SMSendpoint.service

# Prepare filesystem
echo 'Creating installation directory at $INSTALLATION_PATH'
mkdir -p $INSTALLATION_PATH
cd $INSTALLATION_PATH

# Download binary file
echo "Downloading binary"
# TODO: gist on github as Juan suggested
# wget

# Download service file
echo "Downloading service files"
wget https://raw.githubusercontent.com/netsec-ethz/2SMS/master/endpoint/2SMSendpoint.service
# TODO: Modify service file with correct SCION address and IP parameters
sudo mv SERVICE_FILE /etc/systemd/system/
# TODO: chown, chgrp, chmod?

# Check system
[ ! -v $SC ] || [ ! -d $SC ] && { echo "SCION environment variable $SC not set, is SCION installed correctly?"; exit 1 }
[ ! -s $DEFAUL_SOCKET ] && { echo 'SCION default socket not found at $DEFAULT_SOCKET'; exit 1 }

# Start service
sudo systemctl start $SERVICE_FILE

# Check service status
# TODO: check output or status code from service
if [ ]; then
	echo "Successfully installed and started 2SMS Endpoint application"
else; then
	echo 'Failed starting 2SMS Endpoint application, something went wrong during installation'
	exit 1
fi
