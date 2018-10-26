**List Targets**
----
  Return the configured monitoring targets.

* **URL**

  /targets

* **Method:**

  `GET`

* **Success Response:**
  
  * **Code:** 200 <br />
    **Content:** 
    
        [{
            Name: string
            ISD: string
            AS: string
            IP: string
            Port: string
            Path: string
            Labels: {string:string}
    	}]
 
* **Error Response:**

  * **Code:** 500 SERVER ERROR <br />

* **Sample Call:**

  curl -X GET http://127.0.0.1:9999/targets 

* **Notes:**


**Add Target**
----
  Adds a new monitoring target to the configuration.

* **URL**

  /targets

* **Method:**

  `POST`

* **Data Params**

  **Required:**
  
      {
          Name: string
          ISD: string
          AS: string
          IP: string
          Port: string
          Path: string
          Labels: {string:string}
      }

* **Success Response:**
  
  * **Code:** 201 <br />
 
* **Error Response:**

  * **Code:** 500 SERVER ERROR <br />
  
  OR

  * **Code:** 400 <br />

* **Sample Call:**

  curl -X POST http://127.0.0.1:9998/targets -H "Content-Type: application/json" -d '{"Name":"br", "ISD":"11", "AS":"ffaa:0:11", "IP":"127.0.0.1", "Port":"33333", "Path":"/br"}'

* **Notes:**


**Remove Target**
----
  Removes a monitoring target from the configuration.

* **URL**

  /targets

* **Method:**

  `DELETE`

* **Data Params**

  **Required:**
  
      {
          Name: string
          ISD: string
          AS: string
          IP: string
          Port: string
          Path: string
          Labels: {string:string}
      }

* **Success Response:**
  
  * **Code:** 204 <br />
 
* **Error Response:**

  * **Code:** 500 SERVER ERROR <br />
  
  OR

  * **Code:** 400 <br />

* **Sample Call:**

  curl -X DELETE http://127.0.0.1:9998/targets -H "Content-Type: application/json" -d '{"Name":"br", "ISD":"11", "AS":"ffaa:0:11", "IP":"127.0.0.1", "Port":"33333", "Path":"/br"}'

* **Notes:**

**List Storages**
----
  Return the configured remote storages.

* **URL**

  /storages

* **Method:**

  `GET`

* **Success Response:**
  
  * **Code:** 200 <br />
    **Content:** 
    
        [{
            IA:  string
            IP: string
            Port: string
            ManagePort: string -> empty string
    	}]
 
* **Error Response:**

  * **Code:** 500 SERVER ERROR <br />

* **Sample Call:**

  curl -X GET http://127.0.0.1:9999/storages

* **Notes:**


**Add Storage**
----
  Adds a new remote storage to the configuration.

* **URL**

  /storages

* **Method:**

  `POST`

* **Data Params**

  **Required:**
  
      {
            IA:  string
            IP: string
            Port: string
      }

* **Success Response:**
  
  * **Code:** 201 <br />
 
* **Error Response:**

  * **Code:** 500 SERVER ERROR <br />
  
  OR

  * **Code:** 400 <br />

* **Sample Call:**

  curl -X POST http://127.0.0.1:9998/storages -H "Content-Type: application/json" -d '{"IA": "11-ffaa:1:11", "IP": "127.0.0.3", "Port": "8185"}'

* **Notes:**


**Remove Storage**
----
  Removes a remote storage from the configuration.

* **URL**

  /storages

* **Method:**

  `DELETE`

* **Data Params**

  **Required:**
  
      {
            IA:  string
            IP: string
            Port: string
      }

* **Success Response:**
  
  * **Code:** 204 <br />
 
* **Error Response:**

  * **Code:** 500 SERVER ERROR <br />
  
  OR

  * **Code:** 400 <br />

* **Sample Call:**

  curl -X DELETE http://127.0.0.1:9998/storages -H "Content-Type: application/json" -d '{"IA": "11-ffaa:1:11", "IP": "127.0.0.3", "Port": "8185"}'

* **Notes:**
