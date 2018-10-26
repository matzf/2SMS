**Register Endpoint**
----
Adds a new Endpoint to the list of registered Endpoints. Furthermore it automatically adds a target for each
mapping available at the Endpoint to the right scraper(s) and return them.

* **URL**

  /endpoints/register

* **Method:**

  `POST`
  
* **Data Params**

    **Required:**
    
        {
            IA:           string,
            IP:           string,
            ScrapePort:   string,
            ManagePort:   string,
            Paths:        [string]
        }
    
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

  * **Code:** 400 BAD REQUEST <br />

  OR

  * **Code:** 500 SERVER ERROR <br />

* **Sample Call:**

* **Notes:**

  17.08.2018: Add sample call and error messages
  
**Register Scraper**
----
  Adds a new Scraper to the list of registered Scrapers.

* **URL**

  /scrapers/register

* **Method:**

  `POST`
  
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

* **Sample Call:**

* **Notes:**

  17.08.2018: Add sample call and error messages
  
**Notify new Mapping**
----
Signals that a new mapping was added to an Enpoint, i.e. there is a new monitoring target, and will automatically 
add it to the right scraper(s) and return them.

* **URL**

  /endpoint/mappings/notify

* **Method:**

    `POST`

* **Data Params**

        {
            Name: string
            ISD: string
            AS: string
            IP: string
            Port:	string
            Path:	string
            Labels: {string: string}
        }

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

  * **Code:** 400 BAD REQUEST <br />

  OR

  * **Code:** 500 SERVER ERROR <br />

* **Sample Call:**

* **Notes:**

    17.08.2018: Add sample call and error messages
    
**Notify removed Mapping**
----
Signals that a mapping was removed from an Enpoint, i.e. it is not a monitoring target anymore, and will automatically 
remove it from any scraper that has it as a target.

* **URL**

  /endpoint/mappings/notify

* **Method:**

    `DELETE`

* **Data Params**

        {
            Name: string
            ISD: string
            AS: string
            IP: string
            Port:	string
            Path:	string
            Labels: {string: string}
        }

* **Success Response:**
  
  * **Code:** 204 <br />

* **Sample Call:**

* **Notes:**

    17.08.2018: Add sample call