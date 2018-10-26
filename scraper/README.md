## SETTING UP OAUTH2 for Telegraf
Tried by setting env variables and by starting service with options but both didn't work. Following https://www.digitalocean.com/community/tutorials/how-to-monitor-system-metrics-with-the-tick-stack-on-centos-7 and modifying the service file worked.

## Solve prometheus/prometheus and prometheus/common vendoring problem
Problem: using some modules (e.g. LabelSet) in the code causes an error because the version in this project (local path) from the one used in the prometheus project (prometheus vendor package
Similar problem and hacky solution: https://github.com/prometheus/prometheus/issues/1720
Local solution: copied the content of vendor folders into $GOPATH (using cp -rn to not overwrite existing packages) and then removed prometheus/prometheus/vendor and prometheus/common/vendor.
                Another error was then raised (double declaration of a syscall) for golang.org/x/sys/unix/flock.go, removing the file solved the problem.

## Description
TODO: overview of how it works
### Prometheus configuration file
To support SCION addresses without having to change Prometheus itself we add ISD-AS as prefix to any path
in a configuration file. E.g. in a job: `metrics_path: /ISD-AS/metrics`.
Furthermore every request must be proxied through the Scraper's proxies, thus on every job and remote read/write
the option `proxy_url` must be set to point to the respective localhost proxy. E.g. proxy_url: http://127.0.0.1:9901 for
scraping or `proxy_url: http://127.0.0.1:9902 for reading/writing`.

### SCION vs HTTPS server port numbers
It is assumed that the given some port for the SCION server to listen on, the port where the HTTPS server listens
is just one higher. E.g. if an Endpoint has address ISD-AS,[IP]:Port the HTTPS should have address IP:(Port + 1).

## Options
TODO: write options list

## REST API

### Add Target
Add a new scrape target with the given parameters.
#### Request
POST /targets application/json
{   "Name":     string,
    "ISD":      string,
    "AS":       string,
    "IP":       string,
    "Port":     string,
    "Path":     string,
    "Labels":   {string:string}
}
#### Response
204
NoBody
TODO: errors
#### Example
curl -H "Content-Type: application/json" -X POST http://127.0.0.1:9998/targets -d '{"Name":"test", "ISD":"17", "AS":"ffaa:1:41", "IP":"10.0.2.15", "Port":"33333", "Path":"/node"}'

### Remove Target
Removes the given scrape target. Some fields are not needed for matching the target in the list.
#### Request
DELETE /targets application/json
{  "Name": string,
    "ISD": string,
    "AS": string,
    "IP": string,
    "Port": string,             -> not needed
    "Path": string,             -> not needed
    "Labels": {string:string}   -> not needed
}
#### Response
204
NoBody
TODO: errors

### List Targets
Returns a list of all currently configured scrape targets.
#### Request
GET /targets
NoBody

#### Response
200 application/json
[{  "Name": string,
    "ISD": string,
    "AS": string,
    "IP": string,
    "Port": string,
    "Path": string,
    "Labels": {string:string}
}]
TODO: errors
