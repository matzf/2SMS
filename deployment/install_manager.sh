#!/bin/bash

set -e # fail if unexpected error

echo "Started installation of 2SMS Manager application"

# Defaults
INSTALLATION_PATH="$HOME/2SMS/deployment/manager"
DEFAULT_SOCKET=/run/shm/sciond/default.sock
SERVICE_FILE_NAME=2SMSmanager.service
SERVICE_FILE_LOCATION=/etc/systemd/system
IP='192.33.93.196'
monitoring_download_page="monitoring.scionlab.org/downloads/public/manager"
[ -f $SC/gen/ia ] && IA=$(cat $SC/gen/ia | sed 's/_/:/g') || { echo "Missing $SC/gen/ia file"; exit 1; }

# Prepare filesystem
echo "Creating installation directory at $INSTALLATION_PATH"
mkdir -p $INSTALLATION_PATH
cd $INSTALLATION_PATH

echo "Downloading manager binary"
rm -f manager
wget "https://$monitoring_download_page/manager.tar.gz" --quiet
tar -xzvf manager.tar.gz
rm -f manager.tar.gz
chmod +x manager

echo "Downloading configuration files"
wget "https://$monitoring_download_page/manager_configuration.tar.gz" --quiet
tar -xzvf manager_configuration.tar.gz
rm -f manager_configuration.tar.gz
mv -n configuration/* .
rm -rf configuration

# Check contents
if [ ! -f 2SMSmanager.service ]; then
    echo "Some required configuration file is missing from the unpacked manager_configuration archive. Please make sure that you have the following files: 2SMSmanager.service"
    exit 1
fi

# Modify service file with correct SCION address and IP parameters
sed -i -r "s/_USER_/$USER/g;s/_IA_/$IA/g;s/_IP_/$IP/g;" $SERVICE_FILE_NAME

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
    echo "SCION manager's cert directory not found. Aborting."
    exit 1
elif [ ! -S $DEFAULT_SOCKET ]; then
    echo "SCION default socket not found at $DEFAULT_SOCKET"
    exit 1
else
    echo "SCION correctly installed"
fi

# Copy AS certificate
mkdir -p ca_certs
certs_dir=$(find $SC/gen/ -type d -regex ".*/ISD.*/AS.*/endhost/certs")
cp $certs_dir/*.crt ca_certs

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
if systemctl is-active --quiet 2SMSmanager.service; then
    echo "Successfully installed and started 2SMS Manager application"
else
    echo 'Failed starting 2SMS Manager application, something went wrong during installation. Check journal and syslog to have more details.'
    exit 1
fi
