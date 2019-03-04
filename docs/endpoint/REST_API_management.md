**List Mappings**
----
  Returns the configured mappings from paths to local ports.
  
* **URL**

  /mappings

* **Method:**

  `GET`

* **Success Response:**
  
  * **Code:** 200 <br />
    **Content:**
    
        {
            string:string
        }
 
* **Error Response:**

  * **Code:** 500 SERVER ERROR <br />

* **Sample Call:**

  curl -X GET http://127.0.0.1:9999/mappings 

* **Notes:**

**Add Mapping**
----
  Adds a new path to local port mapping to the Endpoint. If a Manager is configured the Target corresponding to the
  new Mapping will be notified and authorization will be granted to any Scraper that adds the new Target.

* **URL**

  /mappings

* **Method:**

  `POST`
  
* **Data Params**

  **Required:**
  
      {
        Path: string,
        Port: string
      }

* **Success Response:**
  
  * **Code:** 201 <br />
 
* **Error Response:**

  * **Code:** 400 BAD REQUEST <br />

  OR

  * **Code:** 500 SERVER ERROR <br />  

* **Sample Call:**

  curl -X POST http://127.0.0.1:9999/mappings -H "Content-Type: application/json" -d '{"Path": "/br", "Port": "32042"}'
  
* **Notes:**

**Remove Mapping**
----
  Removes an existing path to local port mapping from the Endpoint. If a Manager is configured the Target corresponding to the
  removed Mapping will be notified and authorization will removed for any Scraper that has it as Target.

* **URL**

  /mappings

* **Method:**

  `DELETE`
  
* **Data Params**

  **Required:**
  
      {
        Path: string,
        Port: string
      }

* **Success Response:**
  
  * **Code:** 204 <br />
 
* **Error Response:**

  * **Code:** 400 BAD REQUEST <br />

  OR

  * **Code:** 500 SERVER ERROR <br />  

* **Sample Call:**

  curl -X DELETE http://127.0.0.1:9999/mappings -H "Content-Type: application/json" -d '{"Path": "/br", "Port": "32042"}'
  
* **Notes:**
  
**List Mapping's Metrics**
----
  Returns information (Name, Type, Help) about every metric that is exposed at the given Mapping.
  
* **URL**

  /:mapping/metrics/list

* **Method:**

  `GET`
  
*  **URL Params**

   **Required:**
 
   `mapping=string`

* **Success Response:**
  
  * **Code:** 200 <br />
    **Content:** 
    
        [{
                Name: string,
                Type: string,
                Help: string
        }]
 
* **Error Response:**

  * **Code:** 400 BAD REQUEST <br />

  OR

  * **Code:** 500 SERVER ERROR <br />

* **Sample Call:**

  curl -X GET http://127.0.0.1:9999/br/metrics/list
  
* **Notes:**


**Enable Access Control**
----
  Turns on access control for metrics collection.

* **URL**

  /access_control

* **Method:**

  `POST`

* **Success Response:**
  
  * **Code:** 204 <br />

* **Sample Call:**

  curl -X POST http://127.0.0.1:9999/access_control

* **Notes:**

**Disable Access Control**
----
  Turns off access control for metrics collection.

* **URL**

  /access_control

* **Method:**

  `DELETE`

* **Success Response:**
  
  * **Code:** 204 <br />

* **Sample Call:**

  curl -X DELETE http://127.0.0.1:9999/access_control

* **Notes:**

**List Sources**
----
  Returns all the sources that currently are in the authorization policy.
  
* **URL**

  /sources

* **Method:**

  `GET`

* **Success Response:**
  
  * **Code:** 200 <br />
    **Content:**  `[string]`
 
* **Error Response:**

  * **Code:** 500 SERVER ERROR <br />

* **Sample Call:**

  curl -X GET http://127.0.0.1/sources
  
* **Notes:**

**List Source Roles**
----
  Returns all the roles that are assigned to a source.
  
* **URL**

  /:source/roles

* **Method:**

  `GET`
  
*  **URL Params**

   **Required:**
 
   `source=string`
   
* **Success Response:**
  
  * **Code:** 200 <br />
    **Content:**  `[string]`
 
* **Error Response:**

  * **Code:** 500 SERVER ERROR <br />

* **Sample Call:**

  curl -X GET http://127.0.0.1/11-ffaa:0:11/roles
  
* **Notes:**

**Add Source Role**
----
  Adds a role to a source.

* **URL**

  /:source/roles

* **Method:**

  `POST`
  
*  **URL Params**

   **Required:**
 
   `source=string`

* **Data Params**

  **Required:** `string`

* **Success Response:**
  
  * **Code:** 204 <br />
 
* **Error Response:**

  * **Code:** 500 SERVER ERROR <br />

* **Sample Call:**

  curl -H "Content-Type: application/json" -X POST http://127.0.0.1:9999/11-ffaa:0:11/roles -d '"some_role_name"'

* **Notes:**

**Remove Source Role**
----
  Removes a role from a source.

* **URL**

  /:source/roles

* **Method:**

  `POST`
  
*  **URL Params**

   **Required:**
 
   `source=string`

* **Data Params**

  **Required:** `string`

* **Success Response:**
  
  * **Code:** 204 <br />
 
* **Error Response:**

  * **Code:** 500 SERVER ERROR <br />

* **Sample Call:**

  curl -H "Content-Type: application/json" -X DELETE http://127.0.0.1:9999/11-ffaa:0:11/roles -d '"some_role_name"'

* **Notes:**


**List all Source Permissions**
----
  Returns all permissions for a Source. This includes assigned roles as well as temporal permissions.

* **URL**

  /:source/permissions

* **Method:**

  `GET`
  
*  **URL Params**

   **Required:**
 
   `source=string`

* **Success Response:**
  
  * **Code:** 200 <br />
  **Content:** `{string:string}`
 
* **Error Response:**

  * **Code:** 500 SERVER ERROR <br />

* **Sample Call:**

  curl -X GET http://127.0.0.1:9999/11-ffaa:0:11/permissions

* **Notes:**


**Remove all Source Permissions**
----
  Removes all permissions for a Source. This includes assigned roles as well as temporal permissions.

* **URL**

  /:source/permissions

* **Method:**

  `DELETE`
  
*  **URL Params**

   **Required:**
 
   `source=string`

* **Success Response:**
  
  * **Code:** 204 <br />

* **Sample Call:**

  curl -X DELETE http://127.0.0.1:9999/11-ffaa:0:11/permissions

* **Notes:**

**Show Source Status**
----
  Shows a Source's permissions overview with scrape and temporal permissions for each Endpoint's mapping.
  
* **URL**

  /:source/status

* **Method:**

  `GET`
  
*  **URL Params**

   **Required:**
 
    `source=string`

* **Success Response:**
  
  * **Code:** 200 <br />
    **Content:** 
    
        {
            string: {
                        CanScrape: bool
                        Frequency: string
                        Until: string
                    } 
        }
 
* **Error Response:**

  * **Code:** 500 SERVER ERROR <br />

* **Sample Call:**

  curl -X GET http://127.0.0.1:9999/11-ffaa:0:11/status

* **Notes:**


**Block Mapping for Source**
----
  Removes permission for scraping a Mapping for a Source, but doesn't modify temporal permissions or role assignments.
  
* **URL**

  /:source/:mapping/block

* **Method:**

  `GET`
  
*  **URL Params**

   **Required:**
 
    `source=string`
    
    `mapping=string`

* **Success Response:**
  
  * **Code:** 204 <br />

* **Sample Call:**

  curl -X GET http://127.0.0.1:9999/11-ffaa:0:11/br/block

* **Notes:**

**Enable Mapping for Source**
----
  Adds permission for scraping a Mapping for a Source, but doesn't modify temporal permissions or role assignments.
  
* **URL**

  /:source/:mapping/enable

* **Method:**

  `GET`
  
*  **URL Params**

   **Required:**
 
    `source=string`
    
    `mapping=string`

* **Success Response:**
  
  * **Code:** 204 <br />

* **Sample Call:**

  curl -X GET http://127.0.0.1:9999/11-ffaa:0:11/br/enable

* **Notes:**


**Set Mapping's Source Frequency**
----
  Sets the mapping's scraping frequency permission for the source.
  
* **URL**

  /:source/:mapping/frequency

* **Method:**
  
  `POST`
  
*  **URL Params**

   **Required:**
 
    `source=string`
    
    `mapping=string`

* **Data Params**

  **Required:**
   
    `string`, the scrape frequency (e.g. 30s)

* **Success Response:**
  
  * **Code:** 204 <br />

* **Sample Call:**

  curl -X POST http://127.0.0.1:9999/11-ffaa:0:11/br/frequency -H "Content-Type: application/json" -d '"30s""'

* **Notes:**


**Remove Mapping's Source Frequency**
----
  removes the mapping's scraping frequency permission for the source.
  
* **URL**

  /:source/:mapping/frequency

* **Method:**
  
  `DELETE`
  
*  **URL Params**

   **Required:**
 
    `source=string`
    
    `mapping=string`

* **Success Response:**
  
  * **Code:** 204 <br />

* **Sample Call:**

  curl -X DELETE http://127.0.0.1:9999/11-ffaa:0:11/br/frequency

* **Notes:**


**Set Mapping's Time Window**
----
  Sets the mapping's scraping time window permission for the source.
  
* **URL**

  /:source/:mapping/window

* **Method:**
  
  `POST`
  
*  **URL Params**

   **Required:**
 
    `source=string`
    
    `mapping=string`

* **Data Params**

  **Required:**
   
    `string`, the window duration (e.g. 1d)

* **Success Response:**
  
  * **Code:** 204 <br />

* **Sample Call:**

  curl -X POST http://127.0.0.1:9999/11-ffaa:0:11/br/window -H "Content-Type: application/json" -d '"1d""'

* **Notes:**


**Remove Mapping's Time Window**
----
  removes the mapping's scraping time window permission for the source.
  
* **URL**

  /:source/:mapping/window

* **Method:**
  
  `DELETE`
  
*  **URL Params**

   **Required:**
 
    `source=string`
    
    `mapping=string`

* **Success Response:**
  
  * **Code:** 204 <br />

* **Sample Call:**

  curl -X DELETE http://127.0.0.1:9999/11-ffaa:0:11/br/window

* **Notes:**

**List Roles**
----
  Returns a list with all configured roles' name.

* **URL**

  /roles

* **Method:**

  `GET`

* **Success Response:**
  
  * **Code:** 200 <br />
    **Content:** `[string]`
 
* **Error Response:**

  * **Code:** 500 SERVER ERROR <br />

* **Sample Call:**

  curl -X GET http://127.0.0.1:9999/roles
  
* **Notes:**


**Create Role**
----
  Creates a new Role.

* **URL**

  /roles

* **Method:**

  `POST`
  
* **Data Params**

  **Required:**
  
      {
        Name: string,
        Permissions: {string:[string]}
      }

* **Success Response:**
  
  * **Code:** 201 <br />
 
* **Error Response:**

  * **Code:** 400 BAD REQUEST <br />

  OR

  * **Code:** 500 SERVER ERROR <br />

* **Sample Call:**

  curl -X POST http://127.0.0.1:9999/roles -H "Content-Type: application/json" -d '{"Name": "some_role", "Permissions": {"/br": ["metricX", "metricY"], "/cs": ["metricZ"]}'
  
* **Notes:**


**Delete Role**
----
  Deletes a Role.

* **URL**

  /roles/:role

* **Method:**

  `DELETE`
  
*  **URL Params**

   **Required:**
 
   `role=string`, the Role name

* **Success Response:**
  
  * **Code:** 204 <br />

* **Sample Call:**

  curl -X DELETE http://127.0.0.1:9999/roles/some_role
  
* **Notes:**


**Get Role's Info**
----
  Returns information about a Role (basically it's permissions).

* **URL**

  /roles/:role

* **Method:**

  `GET`
  
*  **URL Params**

   **Required:**
 
   `role=string`, the Role name

* **Success Response:**
  
  * **Code:** 200 <br />
  **Content:** 
    
        {
          Name: string,
          Permissions: {string:[string]}
        }

* **Sample Call:**

  curl -X GET http://127.0.0.1:9999/roles/some_role
  
* **Notes:**


**Add Permissions to Role**
----
 Adds permissions to a role for a single Mapping.
 
* **URL**

  /roles/:role/permissions/:mapping

* **Method:**

  `POST`
  
*  **URL Params**
   
   **Required:**
   
    `role=string`
    
    `mapping=string`

* **Data Params**

  **Required:**
  
    `[string]`, Permission names

* **Success Response:**
  
  * **Code:** 204 <br />
 
* **Error Response:**

  * **Code:** 400 BAD REQUEST <br />

  OR

  * **Code:** 500 SERVER ERROR <br />

* **Sample Call:**

  curl -X POST http://127.0.0.1:9999/roles/some_role/permissions/br -H "Content-Type: application/json" -d '["metricX", "metricY"]'
  
* **Notes:**


**Remove Permissions from Role**
----
 Removes permissions from a role for a single Mapping.
 
* **URL**

  /roles/:role/permissions/:mapping

* **Method:**

  `DELETE`
  
*  **URL Params**
   
   **Required:**
   
    `role=string`
    
    `mapping=string`

* **Data Params**

  **Required:**
  
    `[string]`, Permission names

* **Success Response:**
  
  * **Code:** 204 <br />
 
* **Error Response:**

  * **Code:** 400 BAD REQUEST <br />

  OR

  * **Code:** 500 SERVER ERROR <br />

* **Sample Call:**

  curl -X DELETE http://127.0.0.1:9999/roles/some_role/permissions/br -H "Content-Type: application/json" -d '["metricX", "metricY"]'
  
* **Notes:**
