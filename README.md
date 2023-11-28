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

## API
### Login ###
First of all, a user needs to authenticate before sending any kind of information
The simplest way to do it is using third party auth services, like google. There are a lot of libraries and good SDK for Golang.
Login in Google is assumpted to be done in the frontend, therefore an auth token is returned (really a JWT token with user information) once the user is logged, and can be used to verify the identity.
If token is not valid, an error is returned during login process.

** Request **
```json
POST /login/google

{
  "token": "..."}
}
```

** Response OK **
```json
HTTP 200

{
 "user": {
   "name": "John",
   "picture_url": "http://...",
 }
}
```

** Response error **
```json
HTTP 422

{
  "message": "There was an error validating auth token"
}
```

### Send ticket
This is the most important endpoint, because it's the responsible of receiving and processing the ticket information. Despiste being a process which can delay for some seconsd, it's not planned to run as a background service yet.
To process a ticket, an image has to be sent with valid format (currently JPG, TIFF, PDF) and enough quality. If image lacks details or has not enough information, the ticket will not be processed.
Take into consideration each request is charged in AWS, so avoid trial-and-error while uploading tickets.

** Request **
```json
POST /ticket
Content-type: multipart/form-data

...
```

```
** Response ok **
{
  "message": "Ticket created successfully",
  "ticket": {
    "id": 1,
    "date": "1/1/2023",
    "supermarket": "Any",
    "items": [
      {
        "concept": "Bread",
        "quantity": 5
        "unit_price": 1
        "price": 5
      },
      ...
    ]
  }
}
```

** Response error **
```
{
  "message": "Missing file"
}
```

## Frontend
TODO
