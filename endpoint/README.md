## Description
TODO: overview of how it works

## Options
TODO: write options list

## REST API

### Add Mapping
Add a new mapping to redirect traffic for a path to a localhost port and notifies the manager
about the change, so that the new path is added as scrape target.
Important assumption: the path on localhost where the metrics are eyposed is `/metrics`. The mapping
path is only for port mapping purposes.
#### Request
POST /mappings application/json
{   "Path":     string,
    "Port":      string
}
#### Response
200
TODO: errors
#### Example
curl -X POST http://127.0.0.2:9999/mappings {"Path":"/node", "Port":"9100"}

### Remove Mapping
Removes the given mapping and notifies the manager about the change, so that the removed mapping
is removed from the scrape targets.
#### Request
DELETE /mappings application/json
{   "Path":     string,
    "Port":      string
}
#### Response
200
NoBody
TODO: errors

### List Mappings
Returns a list of all currently configured scrape targets.
#### Request
GET /mappings
NoBody

#### Response
200 application/json
[{   "Path":     string,
     "Port":      string
}]
TODO: errors

### List Metrics
#### Request

#### Response
#### Example
curl -X GET http://127.0.0.1:9998/node/metrics/list

### Enable Access Control
#### Request

#### Response
#### Example
curl -X POST http://127.0.0.1:9998/access_control

### Disable Access Control
#### Request

#### Response
#### Example
curl -X DELETE http://127.0.0.1:9998/access_control

### List Source Roles
#### Request

#### Response
#### Example
curl -X GET http://127.0.0.1:9998/test/roles

### Add Source Role
#### Request

#### Response
#### Example
curl -X POST http://127.0.0.1:9998/test/roles -H "Content-Type: application/json" -d '"second"'

### Remove Source Role
#### Request

#### Response
#### Example
curl -X DELETE http://127.0.0.1:9998/test/roles -H "Content-Type: application/json" -d '"second"'

### Source Status
#### Request

#### Response
#### Example
curl -X GET http://127.0.0.1:9998/test/status

### Block Mapping for Source
#### Request

#### Response
#### Example
curl -X GET http://127.0.0.1:9998/test/node/block

### Enable Mapping for Source
#### Request

#### Response
#### Example
curl -X GET http://127.0.0.1:9998/test/node/enable

### Remove Source Frequency
#### Request

#### Response
#### Example
curl -X DELETE http://127.0.0.1:9998/test/node/frequency

### Set Source Frequency
#### Request

#### Response
#### Example
curl -X POST http://127.0.0.1:9998/test/node/frequency -H "Content-Type: application/json" -d '"1m"'

### Remove Source Window
#### Request

#### Response
#### Example
curl -X DELETE http://127.0.0.1:9998/test/node/window


### Set Source Window
#### Request

#### Response
#### Example
curl -X POST http://127.0.0.1:9998/test/node/window -H "Content-Type: application/json" -d '"1d"'

### List Roles
#### Request

#### Response
#### Example
curl -X GET http://127.0.0.1:9998/roles

### Create Role
#### Request

#### Response
#### Example
curl -X POST http://127.0.0.1:9998/roles -H "Content-Type: application/json" -d '{"Name": "ASD", "Permissions": {"/node": ["node_cpu_guest_seconds_total"], "/br": ["scoobydooby"]}}'

### Delete Role
#### Request

#### Response
#### Example
curl -X DELETE http://127.0.0.1:9998/roles/ASD

### Get Role Info
#### Request

#### Response
#### Example
curl -X GET http://127.0.0.1:9998/roles/ASD

### Add Role Permissions
#### Request

#### Response

### Remove Role Permissions
#### Request

#### Response

### List All Sources
#### Request

#### Response
#### Example
curl -X GET http://127.0.0.1:9998/sources