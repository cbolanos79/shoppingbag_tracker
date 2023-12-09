# shoppingbag_tracker
A project focused on tracking the prices in the shopping bag through receipts

## Description
There can not be doubt the shopping bag cost as raised since two years ago, more or less, specially in Spain, where some supermarkets rocketed prices without an apparent reason.  Sometimes it feels like prices change from one week to another, or just in a brief period of time, but given supermarkets don't have public access websites, there is not a way to track online information, at least the most populars in Spain.
The purpose of this project is simple: read prices information from tickets, store and track for changes.
Reading prices from tickets can be done by hand but it would take a very long time, furthermore they can be scanned using AWS Textract, which despiste being a bit expensive, automates a lot of work and allow massive data loading.
Initially, this project is not intended run open to everybody but to run in private instances, to avoid privacity concerns.

## Workflow
As said, AWS Textract is a center piece to scan the tickets, and once scanned, process them: Textract can detect main fields in the ticket and return structured information which can be parsed and stored in a database.
This is the workflow:

- Take a picture of ticket
- Send to API
- Process ticket using Texract
- Extract information and store into database

Once information is stored, it should be easy to group items by concept and check price changes, get the most buyed ones along the month, and recover more useful information.
To achieve these goals, a frontend is required to interact with the final user, which must be very simple and intuitive. Additionally, a backend has to hold the frontend service, like receiving the ticket, sending to Textract, etc
Concerning which stack to use, it seems React and Golang can be a good option: they both have a great community behind and are well known.
For more information about the project, take a look into the Wiki.

## Setup
API and frontend are served from docker images created automatically when docker-compose is run,.
Firstly, create a web/.env file with these variables for frontend:

```
VITE_GOOGLE_CLIENT_ID=...
VITE_API_URL=...
```

Where `VITE_GOOGLE_CLIENT_ID` is the identifier provided by google oauth 2 login, and `VITE_API_URL` is the url where the API is hosted.
It is highly recommended to serve the pages using nginx as reverse proxy with SSL certificates (take a look on Certbot to get a free certificate), and serve API using its own hostname.
Secondly, create a .env file in the project root with these variables for api:

```
GOOGLE_CLIENT_ID=...
JWT_SIGNATURE=
DB_NAME=db
DB_ADAPTER=sqlite3
AWS_ACCESS_KEY_ID=
AWS_SECRET_ACCESS_KEY=
AWS_REGION=
```

Consider these values refer to an IAM user with permissions to use Textract. Be very careful if you are using a 

Finally, take into account react can not load .env files dinamically when build project, therefore .env file for web must be created and ready to use before compiling frontend docker image. It is very important to take into account if .env file is modified, a new docker image must be created.
