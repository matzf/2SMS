#!/bin/bash

set -e # fail if unexpected error

echo "Started installation of 2SMS Scraper application"

# Defaults
INSTALLATION_PATH="$HOME/2SMS/deployment/scraper"
DEFAULT_SOCKET=/run/shm/sciond/default.sock
SERVICE_FILE_NAME=2SMSscraper.service
SERVICE_FILE_LOCATION=/etc/systemd/system
monitoring_download_page="monitoring.scionlab.org/downloads/public/scraper"
### THE FOLLOWING PARAMETERS MAY NEED TO BE ADAPTED TO YOUR SETTING ###
### ------------------------------------------------------ ###
MANAGER_IP='192.33.93.196'
prometheus_retention="10d"
path_prefix="\/prometheus" # Escaped for sed
external_url="https:\/\/monitoring.scionlab.org$path_prefix" # Escaped for sed
prometheus_address="10.10.10.5:9090"
### ------------------------------------------------------ ###
IP=$MANAGER_IP
[ -f $SC/gen/ia ] && IA=$(cat $SC/gen/ia | sed 's/_/:/g') || { echo "Missing $SC/gen/ia file"; exit 1; }

# Prepare filesystem
echo "Creating installation directory at $INSTALLATION_PATH"
mkdir -p $INSTALLATION_PATH
cd $INSTALLATION_PATH

echo "Downloading scraper binary"
rm -f scraper
wget "https://$monitoring_download_page/scraper.tar.gz" --quiet
tar -xzvf scraper.tar.gz
rm -f scraper.tar.gz
chmod +x scraper

echo "Downloading configuration files"
wget "https://$monitoring_download_page/scraper_configuration.tar.gz" --quiet
tar -xzvf scraper_configuration.tar.gz
rm -f scraper_configuration.tar.gz
mv -n configuration/* .
rm -rf configuration

# Check contents
if [ ! -f ca_certs/bootstrap.json ] || [ ! -f ca_certs/ca.crt ] || [ ! -f ca_certs/ISD*AS*.crt ] || [ ! -f prometheus/prometheus.yml ] || [ ! -f prometheus/alert_rules.yml ] || [ ! -f 2SMSscraper.service ]; then
    echo "Some required configuration file is missing from the unpacked scraper_configuration archive. Please make sure that you have the following files: ca_certs/boostrap.json, ca_certs/ca.crt, ca_certs/ISD*AS*.crt, prometheus/prometheus.yml, prometheus/alert_rules.yml, 2SMSscraper.service"
    exit 1
fi

echo "Downloading Prometheus binary"
cd prometheus
rm -f prometheus
wget "https://$monitoring_download_page/prometheus.tar.gz" --quiet
tar -xzvf prometheus.tar.gz
rm -f prometheus.tar.gz
chmod +x prometheus
cd ..

# Modify service file with correct SCION address and IP parameters
sed -i -r "s/_USER_/$USER/g;s/_MANAGER_IP_/$MANAGER_IP/g;s/_IA_/$IA/g;s/_IP_/$IP/g;s/_PROMETHEUS_RETENTION_/$prometheus_retention/g;s/_EXTERNAL_URL_/$external_url/g;s/_PATH_PREFIX_/$path_prefix/g;s/_PROMETHEUS_ADDRESS_/$prometheus_address/g" $SERVICE_FILE_NAME

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
    echo "SCION scraper's cert directory not found. Aborting."
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
if systemctl is-active --quiet 2SMSscraper.service; then
    echo "Successfully installed and started 2SMS Scraper application"
else
    echo 'Failed starting 2SMS Scraper application, something went wrong during installation. Check journal and syslog to have more details.'
    exit 1
fi
