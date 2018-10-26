# Description
TODO: overview of how it works

# Options
TODO: write options list

## REST API 1
This API doesn't require client side verification and offers PKI functionality to the different
system's components.

### Request Certificate
Gets a Certificate Signing Request and after successful validation it and returns the corresponding
certificate signed by the Certificate Authority. If the validation cannot be done automatically (because
the approval of a manager is required) or the CSR is rejected the certificate won't be generated.
If a Certificate for the given CSR was already generated, then it will be just returned.
#### Request
POST /certificate/request application/base64
<base64EncodedCertificateRequest>
#### Response
200
<base64EncodedCertificate>
TODO: errors

### Download Certificate
Checks whether a Certificate for the given IP was generated and in case returns it.
#### Request
GET /certificates/{ip}/get
NoBody

#### Response
200 application/base64
<base64EncodedCertificate>
TODO: errors

## REST API 2
This API requires client side verification and offers PKI and notification functionalities to the
other system's components.

### Register Enpoint
#### Request
POST /endpoints/register application/json
{
    "IA":           string,
    "IP":           string,
    "ScrapePort":   string,
    "ManagePort":   string,
    "Paths":        []string
}
#### Response
204
NoBody
TODO: errors

### Remove Enpoint
TODO: implement
#### Request
DELETE /endpoints/remove application/json
{
    "IA":           string,
    "IP":           string,
    "ScrapePort":   string,
    "ManagePort":   string,
    "Paths":        []string
}
#### Response
TODO

##@ Notify New Mapping
Adds a target for the new mapping (by calling 'Add Target') to all registered scrapers.
#### Request
POST /endpoint/mappings/notify application/json
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
TODO: errors

##@ Notify Removed Mapping
Removes the target for the given mapping (by calling 'Remove Target') from all registered scrapers.
#### Request
DELETE /endpoint/mappings/notify application/json
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

### Register Scraper
#### Request
POST /scrapers/register application/json
{
    "IA":           string,
    "IP":           string,
    "ManagePort":   string,
    "ISDs":        []string
}
#### Response
204
TODO: errors

### Remove Scraper
TODO: implement
#### Request
DELETE /scrapers/remove application/json
{
    "IA":           string,
    "IP":           string,
    "ManagePort":   string,
    "ISDs":        []string
}
#### Response
TODO

## Register Storage
# Request
POST /storages/register application/json application/json
{
     "IA":           string,
     "IP":           string,
     "ManagePort":   string,
     "WritePort":    string
}
# Response
204
NoBody
TODO: errors

## Remove Storage
# Request
DELETE /strorages/remove application/json
{
     "IA":           string,
     "IP":           string,
     "ManagePort":   string,
     "WritePort":    string
}
# Response
204
NoBody
TODO: errors

### REST API 3
This API is meant to be used for managing the system and can be used only over localhost.

### List Certificate Requests
#### Request
GET /manager/certificate/requests
NoBody
#### Response
200 application/json
[x509.CertificateRequest]

### Approve Certificate Request
#### Request
POST /manager/certificate/approve application/json
"<csrFileNameWithoutExtension>"
#### Response
204
NoBody
TODO: errors

### List Registered Enpoints
#### Request
GET /manager/endpoints
NoBody
#### Response
200 application/json
[{
     "IA":           string,
     "IP":           string,
     "ScrapePort":   string,
     "ManagePort":   string,
     "Paths":        []string
}]
TODO: errors

### List Registered Scrapers
#### Request
GET /manager/scrapers
NoBody
#### Response
200 application/json
[{
     "IA":           string,
     "IP":           string,
     "ManagePort":   string,
     "ISDs":        []string
}]

### List Registered Storages
#### Request
GET /manager/storages
NoBody
#### Response
200 application/json
[{
     "IA":           string,
     "IP":           string,
     "ManagePort":   string,
     "WritePort":    string
}]

### List Enpoint's Mappings
#### Request
GET /endpoint/{addr}/mappings
NoBody
#### Response
200 application/json
[{   "Path":     string,
     "Port":      string
}]
TODO: errors

### Add Endpoint Mapping
TODO: implement
#### Request
POST /endpoint/{addr}/mappings application/json
{   "Path":     string,
    "Port":      string
}
#### Response
200
TODO: errors
#### Example
curl -H "Content-Type: application/json" -X POST http://127.0.0.1:10002/endpoint/127.0.0.5:9900/mappings -d '{"Path": "/br", "Port": "32042"}'

### Remove Endpoint Mapping
TODO: implement
#### Request
DELETE /endpoint/{addr}/mappings application/json
{   "Path":     string,
    "Port":      string
}
#### Response
200
NoBody
TODO: errors

### List Scraper's Targets
#### Request
GET /scraper/{addr}/targets
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

### Add Scraper Target
#### Request
POST /scraper/{addr}/targets application/json
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

### Remove Scraper Target
#### Request
DELETE /scraper/{addr}/targets application/json
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
curl -X GET http://127.0.0.1:10002/endpoint/127.0.0.5:9900/roles

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