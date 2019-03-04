**Block Signing**
----
  Disables signing of new certificates, i.e. any new Certificate Request will be negated.
  
* **URL**

  /manager/signing/block

* **Method:**

  `GET`

* **Success Response:**
  
  * **Code:** 204 <br />

* **Sample Call:**

  curl -X GET http://127.0.0.1:10002/manager/signing/block
  
* **Notes:**

**Enable Signing**
----
  Enables signing of new certificates, i.e. any new Certificate Request will be approved, for 1 hour.
  
* **URL**

  /manager/signing/enable

* **Method:**

  `GET`

* **Success Response:**
  
  * **Code:** 200 <br />
  **Content:** `{ "Signing enabled for 1h" }`

* **Sample Call:**

  curl -X GET http://127.0.0.1:10002/manager/signing/enable
  
* **Notes:**

**List Registered Endpoints**
----
  Returns the list of all Endpoints that are currently registered at the Manager.

* **URL**

  /manager/endpoints

* **Method:**

  `GET`

* **Success Response:**
  
  * **Code:** 200 <br />
    **Content:**
    
            [{
                IA:           string,
                IP:           string,
                ScrapePort:   string,
                ManagePort:   string,
                Paths:        [string]
            }]
 
* **Error Response:**

  * **Code:** 500 SERVER ERROR <br />

* **Sample Call:**

  curl -X GET http://127.0.0.1:10002/manager/endpoints
  
* **Notes:**

**List Registered Scrapers**
----
  Returns the list of all Scrapers that are currently registered at the Manager.

* **URL**

  /manager/scrapers

* **Method:**

  `GET`

* **Success Response:**
  
  * **Code:** 200 <br />
    **Content:**
    
        [{ 
            IA: string, 
            IP: string, 
            ManagePort: string, 
            ISDs: [string] 
        }]
 
* **Error Response:**

  * **Code:** 500 SERVER ERROR <br />

* **Sample Call:**

  curl -X GET http://127.0.0.1:10002/manager/scrapers
  
* **Notes:**

**Remove Scraper**
----
  Removes a scraper from the registered Scrapers and removes permissions for each of its Targets.

* **URL**

  /manager/scrapers/remove

* **Method:**

  `DELETE`

* **Data Params**

  **Required:**
  
        [{ 
            IA: string, 
            IP: string, 
            ManagePort: string, 
            ISDs: [string] 
        }]

* **Success Response:**
  
  * **Code:** 204 <br />
 
* **Error Response:**

  * **Code:** 400 BAD REQUEST <br />

  OR

  * **Code:** 500 SERVER ERROR <br />

* **Sample Call:**

  curl -H "Content-Type: application/json" -X DELETE http://127.0.0.1:10002/manager/scrapers/remove -d '{"IA": "17-ffaa:1:43", "IP": "127.0.0.2", "ManagePort": "9900", "ISDs": "17"}'

* **Notes:**

  19.08.2018: Add error messages and what fields are really needed in the data section
  
**Remove Endpoint**
----
  Removes a scraper from the registered Scrapers and removes permissions for each of its Targets.

* **URL**

  /manager/endpoints/remove

* **Method:**

  `DELETE`

* **Data Params**

  **Required:**
  
        [{
            IA:           string,
            IP:           string,
            ScrapePort:   string,
            ManagePort:   string,
            Paths:        [string]
        }]

* **Success Response:**
  
  * **Code:** 204 <br />
 
* **Error Response:**

  * **Code:** 400 BAD REQUEST <br />

  OR

  * **Code:** 500 SERVER ERROR <br />

* **Sample Call:**

  curl -H "Content-Type: application/json" -X DELETE http://127.0.0.1:10002/manager/endpoints/remove -d '{"IA": "17-ffaa:1:43", "IP": "127.0.0.5", "ScrapePort": "9199", "ManagePort": "9900", "Paths": ["/node", "/br"]}'

* **Notes:**

  19.08.2018: Add error messages and what fields are really needed in the data section


**List Endpoint Mappings**
----
  List the mappings that are available at some registered Endpoint by redirecting the call to the Endpoint's API (List Mappings).

* **URL**

  /endpoint/:addr/mappings

* **Method:**

  `GET`
  
*  **URL Params**

   **Required:**
 
   `addr=string`, <IPV4:Port> address of the Endpoint

* **Success Response:**
  
  See Endpoint's API
 
* **Error Response:**

  See Endpoint's API

* **Sample Call:**

  curl -X GET http://127.0.0.1:10002/endpoint/127.0.0.5:9900/mappings
  
* **Notes:**

**Add Endpoint Mapping**
----
  Add a mapping to some registered Endpoint by redirecting the call to the Endpoint's API (Add Mapping).

* **URL**

  /endpoint/:addr/mappings

* **Method:**

  `POST`
  
*  **URL Params**

   **Required:**
 
   `addr=string`, <IPV4:Port> address of the Endpoint

* **Success Response:**
  
  See Endpoint's API
 
* **Error Response:**

  See Endpoint's API

* **Sample Call:**

  curl -X POST http://127.0.0.1:10002/endpoint/127.0.0.5:9900/mappings -H "Content-Type: application/json" -d '{"Path": "/br", "Port": "32042"}'
  
* **Notes:**


**Remove Endpoint Mapping**
----
  Remove a mapping from some registered Endpoint by redirecting the call to the Endpoint's API (Remove Mapping).

* **URL**

  /endpoint/:addr/mappings

* **Method:**

  `DELETE`
  
*  **URL Params**

   **Required:**
 
   `addr=string`, <IPV4:Port> address of the Endpoint

* **Success Response:**
  
  See Endpoint's API
 
* **Error Response:**

  See Endpoint's API

* **Sample Call:**

  curl -X DELETE http://127.0.0.1:10002/endpoint/127.0.0.5:9900/mappings -H "Content-Type: application/json" -d '{"Path": "/br", "Port": "32042"}'
  
* **Notes:**


**List Endpoint Mapping's Metrics**
----
  Returns information (Name, Type, Help) about every metric that is exposed at the given Endpoint's Mapping by by redirecting the call to the Endpoint's API (List Mapping's Metrics).
  
* **URL**

  /endpoint/:addr/:mapping/metrics/list

* **Method:**

  `GET`
  
*  **URL Params**

   **Required:**
 
   `mapping=string`
   
   `addr=string`, <IPV4:Port> address of the Endpoint

* **Success Response:**
  
  See Endpoint's API
 
* **Error Response:**

  See Endpoint's API

* **Sample Call:**

  curl -X GET http://127.0.0.1:10002/endpoint/127.0.0.5:9900/br/metrics/list
  
* **Notes:**


**Enable Endpoint Access Control**
----
  Turns on access control for metrics collection at the Endpoint by redirecting the call to the Endpoint's API (Enable Access Control).

* **URL**

  /endpoint/:addr/access_control

* **Method:**

  `POST`
  
*  **URL Params**

   **Required:**
   
   `addr=string`, <IPV4:Port> address of the Endpoint
   
* **Success Response:**
  
  See Endpoint's API

* **Sample Call:**

  curl -X POST http://127.0.0.1:10002/endpoint/127.0.0.5:9900/access_control

* **Notes:**

**Disable Endpoint Access Control**
----
  Turns off access control for metrics collection at the Endpoint by redirecting the call to the Endpoint's API (Disable Access Control).

* **URL**

  /endpoint/:addr/access_control

* **Method:**

  `DELETE`

*  **URL Params**

   **Required:**
   
   `addr=string`, <IPV4:Port> address of the Endpoint
   
* **Success Response:**
  
  See Endpoint's API

* **Sample Call:**

  curl -X DELETE http://127.0.0.1:10002/endpoint/127.0.0.5:9900/access_control

* **Notes:**

**List Endpoint Sources**
----
  Returns all the sources that currently are in the authorization policy at the Endpoint by redirecting the call to the Endpoint's API (List Sources).
  
* **URL**

  /endpoint/:addr/sources

* **Method:**

  `GET`

*  **URL Params**

   **Required:**
   
   `addr=string`, <IPV4:Port> address of the Endpoint
   
* **Success Response:**
  
  See Endpoint's API
 
* **Error Response:**

  See Endpoint's API

* **Sample Call:**

  curl -X GET http://127.0.0.1:10002/endpoint/127.0.0.5:9900/sources
  
* **Notes:**

**List Endpoint Source Roles**
----
  Returns all the roles that are assigned to a source at the Endpoint by redirecting the call to the Endpoint's API (List Source Roles).
  
* **URL**

   /endpoint/:addr/:source/roles

* **Method:**

  `GET`
  
*  **URL Params**

   **Required:**
 
   `source=string`
   
   `addr=string`, <IPV4:Port> address of the Endpoint
   
* **Success Response:**
  
  See Endpoint's API
 
* **Error Response:**

  See Endpoint's API

* **Sample Call:**

  curl -X GET http://127.0.0.1:10002/endpoint/127.0.0.5:9900/11-ffaa:0:11/roles
  
* **Notes:**

**Add Endpoint Source Role**
----
  Adds a role to a source at the Endpoint by redirecting the call to the Endpoint's API (Add Source Role).

* **URL**

  /endpoint/:addr/:source/roles

* **Method:**

  `POST`
  
*  **URL Params**

   **Required:**
 
   `source=string`
   
   `addr=string`, <IPV4:Port> address of the Endpoint

* **Data Params**

  See Endpoint's API

* **Success Response:**
  
  See Endpoint's API
 
* **Error Response:**

  See Endpoint's API

* **Sample Call:**

  curl -H "Content-Type: application/json" -X POST http://127.0.0.1:10002/endpoint/127.0.0.5:9900/11-ffaa:0:11/roles -d '"some_role_name"'

* **Notes:**

**Remove Endpoint Source Role**
----
  Removes a role from a source at the Endpoint by redirecting the call to the Endpoint's API (Remove Source Role).

* **URL**

  /endpoint/:addr/:source/roles

* **Method:**

  `POST`
  
*  **URL Params**

   **Required:**
 
   `source=string`
   
   `addr=string`, <IPV4:Port> address of the Endpoint

* **Data Params**

  See Endpoint's API

* **Success Response:**
  
  See Endpoint's API
 
* **Error Response:**

  See Endpoint's API

* **Sample Call:**

  curl -H "Content-Type: application/json" -X DELETE http://127.0.0.1:10002/endpoint/127.0.0.5:9900/11-ffaa:0:11/roles -d '"some_role_name"'

* **Notes:**


**List all Endpoint Source Permissions**
----
  Returns all permissions for a Source at the Endpoint by redirecting the call to the Endpoint's API (List all Source Permissions). 
  This includes assigned roles as well as temporal permissions.

* **URL**

  /endpoint/:addr/:source/permissions

* **Method:**

  `GET`
  
*  **URL Params**

   **Required:**
 
   `source=string`
   
   `addr=string`, <IPV4:Port> address of the Endpoint

* **Success Response:**
  
  See Endpoint's API
 
* **Error Response:**

  See Endpoint's API

* **Sample Call:**

  curl -X GET http://127.0.0.1:10002/endpoint/127.0.0.5:9900/11-ffaa:0:11/permissions

* **Notes:**


**Remove all Endpoint Source Permissions**
----
  Removes all permissions for a Source at the Endpoint by redirecting the call to the Endpoint's API (Remove all Source Permissions).
  This includes assigned roles as well as temporal permissions.

* **URL**

  /endpoint/:addr/:source/permissions

* **Method:**

  `DELETE`
  
*  **URL Params**

   **Required:**
 
   `source=string`
   
   `addr=string`, <IPV4:Port> address of the Endpoint

* **Success Response:**
  
  See Endpoint's API

* **Sample Call:**

  curl -X DELETE http://127.0.0.1:10002/endpoint/127.0.0.5:9900/11-ffaa:0:11/permissions

* **Notes:**

**Show Endpoint Source Status**
----
  Shows a Source's permissions overview with scrape and temporal permissions for each Endpoint's mapping at the Endpoint 
  by redirecting the call to the Endpoint's API (Show Source Status).
  
* **URL**

  /endpoint/:addr/:source/status

* **Method:**

  `GET`
  
*  **URL Params**

   **Required:**
 
    `source=string`
    
    `addr=string`, <IPV4:Port> address of the Endpoint

* **Success Response:**
  
  See Endpoint's API
 
* **Error Response:**

  See Endpoint's API

* **Sample Call:**

  curl -X GET http://127.0.0.1:10002/endpoint/127.0.0.5:9900/11-ffaa:0:11/status

* **Notes:**


**Block Endpoint Mapping for Source**
----
  Removes permission for scraping a Mapping for a Source at the Endpoint by redirecting the call to the Endpoint's API (Block Mapping for Source), but doesn't modify temporal permissions or role assignments.
  
* **URL**

  /endpoint/:addr/:source/:mapping/block

* **Method:**

  `GET`
  
*  **URL Params**

   **Required:**
 
    `source=string`
    
    `mapping=string`
    
    `addr=string`, <IPV4:Port> address of the Endpoint

* **Success Response:**
  
  See Endpoint's API

* **Sample Call:**

  curl -X GET http://127.0.0.1:10002/endpoint/127.0.0.5:9900/11-ffaa:0:11/br/block

* **Notes:**

**Enable Endpoint Mapping for Source**
----
  Adds permission for scraping a Mapping for a Source at the Endpoint by redirecting the call to the Endpoint's API (Enable Mapping for Source), but doesn't modify temporal permissions or role assignments.
  
* **URL**

  /endpoint/:addr/:source/:mapping/block

* **Method:**

  `GET`
  
*  **URL Params**

   **Required:**
 
    `source=string`
    
    `mapping=string`
    
    `addr=string`, <IPV4:Port> address of the Endpoint

* **Success Response:**
  
  See Endpoint's API

* **Sample Call:**

  curl -X GET http://127.0.0.1:10002/endpoint/127.0.0.5:9900/11-ffaa:0:11/br/enable

* **Notes:**


**Set Endpoint Mapping's Source Frequency**
----
  Sets the mapping's scraping frequency permission for the source at the Endpoint by redirecting the call to the Endpoint's API (Set Mapping's Source Frequency).
  
* **URL**

  /endpoint/:addr/:source/:mapping/frequency

* **Method:**
  
  `POST`
  
*  **URL Params**

   **Required:**
 
    `source=string`
    
    `mapping=string`
    
    `addr=string`, <IPV4:Port> address of the Endpoint

* **Data Params**

  **Required:**
   
    `string`, the scrape frequency (e.g. 30s)

* **Success Response:**
  
  See Endpoint's API

* **Sample Call:**

  curl -X POST http://127.0.0.1:10002/endpoint/127.0.0.5:9900/11-ffaa:0:11/br/frequency -H "Content-Type: application/json" -d '"30s""'

* **Notes:**


**Remove Endpoint Mapping's Source Frequency**
----
  removes the mapping's scraping frequency permission for the source at the Endpoint by redirecting the call to the Endpoint's API (Remove Mapping's Source Frequency).
  
* **URL**

  /endpoint/:addr/:source/:mapping/frequency

* **Method:**
  
  `DELETE`
  
*  **URL Params**

   **Required:**
 
    `source=string`
    
    `mapping=string`
    
    `addr=string`, <IPV4:Port> address of the Endpoint

* **Success Response:**
  
  See Endpoint's API

* **Sample Call:**

  curl -X DELETE http://127.0.0.1:10002/endpoint/127.0.0.5:9900/11-ffaa:0:11/br/frequency

* **Notes:**


**Set Endpoint Mapping's Time Window**
----
  Sets the mapping's scraping time window permission for the source at the Endpoint by redirecting the call to the Endpoint's API (Set Mapping's Time Window).
  
* **URL**

  /endpoint/:addr/:source/:mapping/window

* **Method:**
  
  `POST`
  
*  **URL Params**

   **Required:**
 
    `source=string`
    
    `mapping=string`
    
    `addr=string`, <IPV4:Port> address of the Endpoint

* **Data Params**

  **Required:**
   
    `string`, the window duration (e.g. 1d)

* **Success Response:**
  
  See Endpoint's API

* **Sample Call:**

  curl -X POST http://127.0.0.1:10002/endpoint/127.0.0.5:9900/11-ffaa:0:11/br/window -H "Content-Type: application/json" -d '"1d""'

* **Notes:**


**Remove Endopint Mapping's Time Window**
----
  removes the mapping's scraping time window permission for the source at the Endpoint by redirecting the call to the Endpoint's API (Remove Mapping's Time Window).
  
* **URL**

  /endpoint/:addr/:source/:mapping/window

* **Method:**
  
  `DELETE`
  
*  **URL Params**

   **Required:**
 
    `source=string`
    
    `mapping=string`
    
    `addr=string`, <IPV4:Port> address of the Endpoint

* **Success Response:**
  
  See Endpoint's API

* **Sample Call:**

  curl -X DELETE http://127.0.0.1:10002/endpoint/127.0.0.5:9900/11-ffaa:0:11/br/window

* **Notes:**

**List Roles at Endpoint**
----
  Returns a list with all configured roles' name at the Endpoint by redirecting the call to the Endpoint's API (List Roles).

* **URL**

  /endpoint/:addr/roles

* **Method:**

  `GET`
  
*  **URL Params**

     **Required:**
      
      `addr=string`, <IPV4:Port> address of the Endpoint

* **Success Response:**
  
  See Endpoint's API
 
* **Error Response:**

  See Endpoint's API

* **Sample Call:**

  curl -X GET http://127.0.0.1:10002/endpoint/127.0.0.5:9900/roles
  
* **Notes:**


**Create Role at Endpoint**
----
  Creates a new Role at the Endpoint by redirecting the call to the Endpoint's API (Create Role).

* **URL**

  /endpoint/:addr/roles

* **Method:**

  `POST`
  
*  **URL Params**

     **Required:**
      
      `addr=string`, <IPV4:Port> address of the Endpoint
  
* **Data Params**

  **Required:**
  
      {
        Name: string,
        Permissions: {string:[string]}
      }

* **Success Response:**
  
  See Endpoint's API
 
* **Error Response:**

  See Endpoint's API

* **Sample Call:**

  curl -X POST http://127.0.0.1:10002/endpoint/127.0.0.5:9900/roles -H "Content-Type: application/json" -d '{"Name": "some_role", "Permissions": {"/br": ["metricX", "metricY"], "/cs": ["metricZ"]}'
  
* **Notes:**


**Delete Role at Endpoint**
----
  Deletes a Role at the Endpoint by redirecting the call to the Endpoint's API (Delete Role).

* **URL**

  /endpoint/:addr/roles/:role

* **Method:**

  `DELETE`
  
*  **URL Params**

   **Required:**
 
   `role=string`, the Role name
   
   `addr=string`, <IPV4:Port> address of the Endpoint

* **Success Response:**
  
  See Endpoint's API

* **Sample Call:**

  curl -X DELETE http://127.0.0.1:10002/endpoint/127.0.0.5:9900/roles/some_role
  
* **Notes:**


**Get Role's Info at Endpoint**
----
  Returns information about a Role (basically it's permissions) at the Endpoint by redirecting the call to the Endpoint's API (Get Role's Info).

* **URL**

  /endpoint/:addr/roles/:role

* **Method:**

  `GET`
  
*  **URL Params**

   **Required:**
 
   `role=string`, the Role name
   
   `addr=string`, <IPV4:Port> address of the Endpoint

* **Success Response:**
  
  See Endpoint's API

* **Sample Call:**

  curl -X GET http://127.0.0.1:10002/endpoint/127.0.0.5:9900/roles/some_role
  
* **Notes:**


**Add Permission to Role at Endpoint**
----
 Adds permissions to a role for a single Mapping at the Endpoint by redirecting the call to the Endpoint's API (Add Permission to Role).
 
* **URL**

  /endpoint/:addr/roles/:role/permissions/:mapping

* **Method:**

  `POST`
  
*  **URL Params**
   
   **Required:**
   
    `role=string`
    
    `mapping=string`
    
    `addr=string`, <IPV4:Port> address of the Endpoint

* **Data Params**

  **Required:**
  
    `[string]`, Permission names

* **Success Response:**
  
  See Endpoint's API
 
* **Error Response:**

  See Endpoint's API

* **Sample Call:**

  curl -X POST http://127.0.0.1:10002/endpoint/127.0.0.5:9900/roles/some_role/permissions/br -H "Content-Type: application/json" -d '["metricX", "metricY"]'
  
* **Notes:**


**Remove Permission from Role at Endpoint**
----
 Removes permissions from a role for a single Mapping at the Endpoint by redirecting the call to the Endpoint's API (Remove Permission from Role).
 
* **URL**

  /endpoint/:addr/roles/:role/permissions/:mapping

* **Method:**

  `DELETE`
  
*  **URL Params**
   
   **Required:**
   
    `role=string`
    
    `mapping=string`
    
    `addr=string`, <IPV4:Port> address of the Endpoint

* **Data Params**

  **Required:**
  
    `[string]`, Permission names

* **Success Response:**
  
  See Endpoint's API
 
* **Error Response:**

  See Endpoint's API

* **Sample Call:**

  curl -X DELETE http://127.0.0.1:10002/endpoint/127.0.0.5:9900/roles/some_role/permissions/br -H "Content-Type: application/json" -d '["metricX", "metricY"]'
  
* **Notes:**


**List Scraper Targets**
----
  Return the configured monitoring targets at the Scraper.

* **URL**

  /scraper/:addr/targets

* **Method:**

  `GET`
  
*  **URL Params**

   **Required:**
   
   `addr=string`, <IPV4:Port> address of the Endpoint
   
* **Success Response:**
  
  See Scraper's API
 
* **Error Response:**

  See Scraper's API

* **Sample Call:**

  curl -X GET http://127.0.0.1:10002/scaper/127.0.0.2:9900/targets 

* **Notes:**


**Add Scraper Target**
----
  Adds a new monitoring target to the scraper's configuration.

* **URL**

  /scraper/:addr/targets

* **Method:**

  `POST`
  
*  **URL Params**

   **Required:**
   
   `addr=string`, <IPV4:Port> address of the Endpoint
   
* **Data Params**

  **Required:**
  
  See Scraper's API

* **Success Response:**
  
  See Scraper's API
 
* **Error Response:**

  See Scraper's API

* **Sample Call:**

  curl -X POST http://127.0.0.1:10002/scaper/127.0.0.2:9900/targets -H "Content-Type: application/json" -d '{"Name":"br", "ISD":"11", "AS":"ffaa:0:11", "IP":"127.0.0.1", "Port":"33333", "Path":"/br"}'

* **Notes:**


**Remove Scraper Target**
----
  Removes a monitoring target from the scraper's configuration.

* **URL**

  /scraper/:addr/targets

* **Method:**

  `DELETE`

* **Data Params**

  **Required:**
  
   See Scraper's API

* **Success Response:**
  
   See Scraper's API
 
* **Error Response:**

  See Scraper's API

* **Sample Call:**

  curl -X DELETE http://127.0.0.1:10002/scaper/127.0.0.2:9900/targets -H "Content-Type: application/json" -d '{"Name":"br", "ISD":"11", "AS":"ffaa:0:11", "IP":"127.0.0.1", "Port":"33333", "Path":"/br"}'

* **Notes:**

**List Scraper Storages**
----
  Return the configured remote storages at the Scraper.

* **URL**

  /scraper/:addr/storages

* **Method:**

  `GET`
  
*  **URL Params**

   **Required:**
   
   `addr=string`, <IPV4:Port> address of the Endpoint
   
* **Success Response:**
  
  See Scraper's API
 
* **Error Response:**

  See Scraper's API

* **Sample Call:**

  curl -X GET http://127.0.0.1:10002/scaper/127.0.0.2:9900/storages 

* **Notes:**


**Add Scraper Storage**
----
  Adds a new remote storage to the scraper's configuration.

* **URL**

  /scraper/:addr/storages

* **Method:**

  `POST`
  
*  **URL Params**

   **Required:**
   
   `addr=string`, <IPV4:Port> address of the Endpoint
   
* **Data Params**

  **Required:**
  
  See Scraper's API

* **Success Response:**
  
  See Scraper's API
 
* **Error Response:**

  See Scraper's API

* **Sample Call:**

  curl -X POST http://127.0.0.1:10002/scaper/127.0.0.2:9900/storages -H "Content-Type: application/json" -d '{"IA": "11-ffaa:1:11", "IP": "127.0.0.3", "Port": "8185", "ManagePort":"9900"}'

* **Notes:**


**Remove Scraper Storage**
----
  Removes a remote storage from the scraper's configuration.

* **URL**

  /scraper/:addr/storages

* **Method:**

  `DELETE`

* **Data Params**

  **Required:**
  
   See Scraper's API

* **Success Response:**
  
   See Scraper's API
 
* **Error Response:**

  See Scraper's API

* **Sample Call:**

  curl -X DELETE http://127.0.0.1:10002/scaper/127.0.0.2:9900/storages -H "Content-Type: application/json" -d '{"IA": "11-ffaa:1:11", "IP": "127.0.0.3", "Port": "8185", "ManagePort":"9900"}'

* **Notes:**