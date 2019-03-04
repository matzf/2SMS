**Request Certificate**
----
  Expects a Certificate Signing Request and, if its validation is successful or the certificate already exists, it returns a corresponding valid certificate signed by the certificate authority.

* **URL**

    /certificate/request

* **Method:**
  
    `POST`

* **Data Params**

    **Required:**
    
  `Base64 encoding of raw x509 Certificate Sign Request`

* **Success Response:**

  * **Code:** 200 <br />
    **Content:** `Base64 encoding of raw x509 Certificate`
 
* **Error Response:**

  * **Code:** 400 BAD REQUEST <br />
  
  OR
    
  * **Code:** 401 UNAUTHORIZED <br />
    **Content:** `{ "Certificate request is blocked. Contact the administrator." }`

* **Sample Call:** 

* **Notes:**

  17.08.2018: be more specific about csr and cert format and create a Sample Call
  
**Download Certificate**
----
  Checks whether a Certificate for the given IP already exists and in case returns it.

* **URL**

  /certificates/:type/:ip/get

* **Method:**

  `GET`
  
*  **URL Params**

   **Required:**
 
   `type=string`, Endpoint | Scraper | Storage
   
   `ip=string`, must be in IPV4 format

* **Success Response:**
  
  * **Code:** 200 <br />
    **Content:** Base64 encoding of raw x509 Certificate.
 
* **Error Response:**

  * **Code:** 404 NOT FOUND <br />
    **Content:** `{ "Certificate doesn't exists" }`

* **Sample Call:**

  curl -X GET https://127.0.0.1:10000/certificates/Scraper/127.0.0.1/get

* **Notes:**

    17.08.2018: be more specific about csr and cert format (same as in Request Certificate)