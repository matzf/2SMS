# 2SMS
SCION Monitoring System

## Project Structure
- common/:       contains modules used in more than one component
- deployment/:   
- docs/:         
- endpoint/:     contains all files related to the Endpoint component
- manager/:      contains all files related to the Manager component
- scraper/:      contains all files related to the Scraper component
- storage/:      contains all files related to the Storage component
- services/:     TODO
- alerting/:     TODO

## Building from source
Make sure that you have [Go 1.9](https://golang.org/dl/) and [SCION](https://netsec-ethz.github.io/scion-tutorials/) installed. 
Then all you need to do in order to build the applications is to run:
- `govendor sync`
- `deps.sh`
- `./build.sh`

in the project's root directory.