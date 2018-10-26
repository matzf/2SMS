# 2SMS
## What is this Project
TODO

## Project Structure
- common:       contains modules used in more than one component
- deployment:   contains binary for each component and a small tutorial
- endpoint:     contains all files related to the Endpoint component
- manager:      contains all files related to the Manager component
- proto:        contains modules aimed to test specific settings and behaviours (should be trashed)
- scraper:      contains all files related to the Scraper component
- storage:      contains all files related to the Storage component

## Deployment Procedure
1. Start a Manager instance
    - requires
2. Start Alertmanager
3. Start Scraper instances (as many as required)
    - requires
    - Point each one to the alertmanager
    - start also Thanos Sidecar process aside from it and build hanos cluster
4. Start Endpoint instances (one per machine to monitor)
    - requires
5. Start Thanos Query instance and connect to cluster
6. Start Grafana Server
    - connect to Query instance

## Thanos Deployment
Following https://github.com/improbable-eng/thanos/blob/master/docs/getting_started.md
### Single Prometheus server
run a prometheus server

run a sidecar attached to the prometheus server
./thanos sidecar --prometheus.url http://localhost:9090 --tsdb.path prometheus-2.2.1.linux-amd64/data/ --grpc-address 0.0.0.0:19091 --http-address 0.0.0.0:19191 --cluster.address 0.0.0.0:19391 --cluster.advertise-address 127.0.0.1:19391 --cluster.peers 127.0.0.1:19391

run a query and point it to the sidecar
./thanos query --http-address 0.0.0.0:19092 --cluster.address 0.0.0.0:19591 --cluster.advertise-address 127.0.0.1:19591 --cluster.peers 127.0.0.1:19391 --query.replica-label replica &> outQuery &

then query http-address

### Cluster with multiple Prometheus servers

run 2 prometheus instances on different IPs
./prometheus --web.listen-address="127.0.1.1:9090"
./prometheus --web.listen-address="127.0.2.1:9090"

run a sidecar for each prometheus instance on the same IP, point the second to the first one to build the cluster (via cluster.peers)
./thanos sidecar --prometheus.url http://127.0.1.1:9090 --tsdb.path prometheus-2.2.1.linux-amd64/data/ --grpc-address 127.0.1.1:19091 --http-address 127.0.1.1:19191 --cluster.address 127.0.1.1:19391 --cluster.advertise-address 127.0.1.1:19391 --cluster.peers 127.0.1.1:19391
./thanos sidecar --prometheus.url http://127.0.2.1:9090 --tsdb.path prometheus2/data/ --grpc-address 127.0.2.1:19091 --http-address 127.0.2.1:19191 --cluster.address 127.0.2.1:19391 --cluster.advertise-address 127.0.2.1:19391 --cluster.peers 127.0.1.1:19391

run a query on another IP and point it to the first sidecar to join the cluster (aagain via cluster peers)
./thanos query --http-address 127.0.0.1:19092 --cluster.address 127.0.0.1:19591 --cluster.advertise-address 127.0.0.1:19591 --cluster.peers 127.0.1.1:19391

query http-address in the browser to see data from both prometheus instances (e.g. up)