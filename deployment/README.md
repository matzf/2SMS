# Deployment of 2SMS
The scripts in this folder can be used to initially deploy the different components of 2SMS. All the scripts can be adapted (or
even be used directly) to perform an update of an existing deployment.
There are some points that should be noted:
* The current system fully supports only a single Scraper instance (multiple instances can be deployed but are not interconnected)
* In the default setup Manager and Scraper are installed in the same local network behind a NAT, other setups are possible but
    require adapting the install scripts (application's parameters)
* If a deployment from scratch is required, then some files in the downloaded configuration archives must be updated (see below for more details)
* Deployment of other external components (e.g. alertmanager, blackbox_exporter, node_exporter, ...) have to be done manually (see below)
* At this stage understanding of Prometheus other than its query language is still necessary as different operations have to
    be performed manually. The documentation can be found at https://prometheus.io/docs/introduction/overview/

## Requirements
* Every instance needs to be reachable via a public IPv4 address
* Every installation scripts will require sudo right to run systemctl commands
* SCION must be installed and a connection to SCIONLab is required
    * $SC/gen/ia must exist
    * /run/shm/sciond/default.sock must exist
    * the AS must be set up as endhost

## Procedure to deploy from scratch
1. Run `install_manager.sh` (change IP accordingly to the setting) to install and run the Manager. Note that this
    will create a new CA with corresponding root certificate and therefore all old issued certificates (e.g. to Endpoints)
    will not be trusted anymore. If you already have a 2SMS installation and wish to keep the same root certificate then
    you need to copy the old ca directory in the new installation path (default to ~/2SMS/deployment/manager)
1. Replace the following files in scraper_configuration/ca_certs and endpoint_configuration/ca_certs with the ones
    that are generated in manager/ca:
        - bootstrap.json
        - ca.crt
   Replace the AS certificate in those two directories with the (latest) one in the manager/ca_certs directory
1. Compress (.tar.gz) the modified directory and load them on the website
1. Run `curl localhost:10002/manager/signing/enable` to enable creation of new certificates for new Scraper or Endpoint instances
    that will be created next (otherwise the component will block and periodically try to get a certificate)
1. Run `install_scraper.sh` (change MANAGER_IP and other parameters accordingly to the setting) to install and run the Scraper.
1. Run `install_endpoint.sh` (change MANAGER_IP accordingly to your setting) in another machine to install and run an Endpoint instance.
1. When you are finished run `curl localhost:10002/manager/signing/block` to block creation of new certificates at the manager

## Procedure to deploy on the existing SCIONLab infrastructure
### Install another Scraper
TODO
### Install another Endpoint
TODO

# Deployment of external components

## Procedure to install Alertmanager
1. Download the binary and the configuration file from `https://monitoring.scionlab.org/downloads/public/alertmanager/`
1. Add the secret tokens to the configuration file
1. Optional: Modify the configuration file. E.g. Add/remove receivers, change routes, use a new Slack channel, ...
1. Download the service file from `https://monitoring.scionlab.org/downloads/public/alertmanager/`
1. Optional: Adapt it to the setting. E.g. change external URL, add parameter to bind to a different port
1. Move the service file to `/etc/systemd/system/`
1. Enable and start the service

## Procedure to install blackbox exporter
**Deprecated**: this will be replaced when issue [#82](https://github.com/netsec-ethz/2SMS/issues/82) is solved.

## Procedure to install node exporter
1. Skip if there is already a node exporter installed and running. Hints showing that this might be the case are:
   - Port 9100 being already used
   - presence of processes or packages with names containing/similar to `node-exporter` (e.g. `prometheus-node-exporter` Debian package)
1. Download the node_exporter binary from `https://monitoring.scionlab.org/downloads/public/node_exporter/`
1. Download the service file from `https://monitoring.scionlab.org/downloads/public/node_exporter/`
1. Optional: Adapt it to the setting. E.g. change the file location.
1. Move the service file to `/etc/systemd/system/`
1. Enable and start the service

## Procedure to install Grafana
From `https://grafana.com/grafana/download`
1. `wget https://dl.grafana.com/oss/release/grafana_6.0.0_amd64.deb`
1. `sudo dpkg -i grafana_6.0.0_amd64.deb`
1. Enable and start `grafana-server.service`
1. Temporary: import current settings in a quick&dirty way by copying grafana sqlite database (`/var/lib/grafana/grafana.db`)
1. Configure Grafana to run behind a proxy by setting the followings in `[server]`:
   - `domain = monitoring.scionlab.org`
   - `root_url = https://monitoring.scionlab.org/new_grafana`
1. Configure Grafana not to request user login by setting the followings in `[auth.anonymous]`:
    - `enabled = true`
    - `org_name = Main Org.`
    - `org_role = Editor`
1. Restart the service
1. Optional: install plugins:
   - new plugins: http://docs.grafana.org/plugins/installation/
   - quick&dirty for old plugins: copy `/var/lib/grafana/plugins`
1. Make sure that data sources are still correct:
   1. login as admin
   1. navigate to `Configuration/Data Sources`
   1. check each data source

## Procedure to install InfluxDB
From `https://portal.influxdata.com/downloads/` (Latest 1.x version, Ubuntu & Debian):
1. `wget https://dl.influxdata.com/influxdb/releases/influxdb_1.7.4_amd64.deb`
1. `sudo dpkg -i influxdb_1.7.4_amd64.deb`
1. Enable and start `influxdb.service`
1. Create `prometheus` database running `influx -execute "create database prometheus"`

# FAQ
## How do I add a new alert?
To add a new alert it's necessary to change `scraper/prometheus/alert_rules.yml`.
A simple alert has this format:

    - alert: <alert_name>               # A unique name for the alert
        expr: <triggering_expression>   # The expression (PromQL syntax) that will trigger the alert when evaluated to true
        for: <duration>                 # For how long the expression has to evaluate to true before the alert is really fired
More information about syntax and options can be found at `https://prometheus.io/docs/prometheus/latest/configuration/alerting_rules/`.

Finally the Prometheus server must be reloaded for the changes to be applied. In a basic installation his can be done with the following command:
`curl -X POST localhost:9090/-/reload`. In a custom installation the URL may be different (e.g. see `reload_prometheus.sh` in PVM).

## How do I change how an alert is notified?
This is the job of the Alertmanager, so we need to modify its configuration file. It is not necessarily required because as long as a
default channel is configured, any (new) alert will be notified over that channel. Detailed information about its structure can
be found in the [official documentation](`https://prometheus.io/docs/alerting/overview/`), so here we provide only a short summary.
TODO

Finally the Alertmanager server must be realoaded for the changes to be applied. In a basic installation his can be done with the following command:
`curl -X POST localhost:9093/-/reload`. In a custom installation the URL may be different (e.g. see `reload_alertmanager.sh` in PVM).

## How is a target configured in the Prometheus' server configuration file?
Every target is configured as a job with this format:
    
    - job_name: <IA> <IP <type>             # A unique name identifying the target (e.g. 17-ffaa:1:c5 127.0.0.1 bs)
      metrics_path: /<IA>/<metrics_path>    # A combination of the target's IA and the path where metrics can be found (e.g. /17-ffaa:1:c5/bs)
      static_configs:                       #
      - targets:                            #
        - <IP>:<Port>                       # Target's address with its IPv4 address and port (e.g. 192.33.93.195:9199)
        labels:                             # List of key-value pairs that can be used to further label the scraped data
          AS: <AS> (e.g. ffaa:1:c5)         #
          ISD: "17"                         #
          service: bs                       #
    proxy_url: http://127.0.0.1:9901        # URL of the Scraper's scraping server (not required for local targets like the Prometheus server itself)

More information about options and syntax can be found at `https://prometheus.io/docs/prometheus/latest/configuration/configuration/`.

## I started a new Endpoint but it's not being monitored, what went wrong?
There are multiple possible reasons for this behaviour:
* Signing of new certificates at the manager is blocked and the Endpoint blocks until it is able to get one
* The Endpoint's targets weren't added to any Scraper:
  * because there is no Scraper responsible for the Endpoint's ISD registered at the Manager
  * because the Manager can't reach any responsible Scraper to add the targets
  * because there was a problem while adding the target's to the Prometheus server
  
Detailed information about the Endpoint process can be found in the syslog file (`/var/log/syslog`) or using journalctl (`journalctl -ru 2SMSendpoint.service`).
The Endpoint logs per default all scrape requests (both successful and failed), while the Scraper logs only failed scrape requests and additions/removals of targets.

## There are false positives, how can I found out why?
This usually happens when there problems in connecting to the Endpoint. By checking the Scraper's output 

## how can I backup the InfluDB data?
Documentation to this purpose can be found [here](https://docs.influxdata.com/influxdb/v1.7/administration/backup_and_restore/).
In short:
- `influxd backup -portable -database prometheus /tmp/influx_prometheus_backup` will create a backup of the database `prometheus` and store the generated files in `/tmp/influx_prometheus_backup`
   - Online backup over TCP connection to the database is possible, for more details see the above mentioned documentation.
- `influxd restore -portable -db prometheus influx_backup_test/` will restore the `prometheus` database with the data found in `influx_backup_test/`
   - Note: if the database already exists an error will be returned, so the database is not simply overwritten. In such a case 
   restore to a temporary database and then sideload the data into the existing DB using `SELECT ... INTO` statement.