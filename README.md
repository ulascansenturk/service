

# Banking Service

This application is a banking service that depends on Temporal, Redis, and PostgreSQL. The required resources are orchestrated using Docker Compose, and the database is automatically migrated once all services are healthy.



## Prerequisites

#### Ensure that Docker and Docker Compose are installed on your machine

Running the Application

 - Clone the repository and change into the project directory:
```
git clone https://github.com/ulascansenturk/service.git && cd service

```
- ```docker-compose up```

This command will:

- Launch the required services (Temporal, Redis, PostgreSQL, API, Temporal-Transfers-Workfer).
- Wait for all services to become healthy.
- Automatically run the database migrations.


Access the Temporal UI once all services are running at:
http://localhost:8233/

The Temporal UI can be used to monitor transfer workflows.

## Making a Money Transfer

To make a money transfer, you first need to create a user. The application will automatically create a bank account for the user in the background.

## Reliability of Transfers

The reliability of money transfers is ensured through the use of a Redis lock before the actual transfer process begins. This locking mechanism prevents race conditions and ensures that only one transfer can occur for a specific source account at a time.

Example Request:

To create a user and their associated bank account, use the following curl command:

```sh
curl --location 'localhost:3000/v1/users' \
--header 'Content-Type: application/json' \
--data-raw '{
    "data":{
        "email":"ulascansenturk1@gmail.com",
        "password": "random-password",
        "firstName":"Ulascan",
        "lastName":"Senturk",
        "currencyCode":"USD",
        "balance":10000
    }
}'
```

Example Response:

```json
{
    "bankAccount": {
        "balance": 10000,
        "currency": "USD",
        "id": "fa8ede6f-8f36-4e71-93ec-4df28af5d0c5",
        "status": "ACTIVE",
        "user_id": "a0e42ad6-474b-4f99-bbe8-4da7ec75ab65"
    },
    "user": {
        "email": "ulascansenturk1@gmail.com",
        "first_name": "Ulascan",
        "id": "a0e42ad6-474b-4f99-bbe8-4da7ec75ab65",
        "last_name": "Senturk"
    }
}

```

To trigger a money transfer between two accounts, use the following curl command:

```sh
curl --location 'localhost:3000/v1/transfers' \
--header 'Content-Type: application/json' \
--data '{
    "data":{
        "reference_id": "7ad62627-2a80-4e62-819e-477802449da4", // Random generated uuid, also workflow id
        "amount":100,
        "fee_amount":230, // Optional
        "sourceAccountID":"068d81de-be46-4d59-a20a-8d3c168504ff",
        "destinationAccountID": "60f4c37f-3509-41a8-b3c1-836c0bb70d39"
    }
}'
```

Example Response:

```json
{
    "SourceTransactionReferenceID": "0d305986-3fde-5792-a867-a625d898e9fb",
    "DestinationTransactionReferenceID": "beecbcc3-bd14-5b6a-93f3-b320694527c4",
    "FeeTransactionReferenceID": "a0f60930-b58d-57b7-bb90-463c9f42e2f2",
    "FeeTransaction": {
        "id": "b11cf134-f3b0-400e-88c6-4902230a24fb",
        "user_id": "3f1060e4-18f2-4b23-bc93-bd04ae29f9e7",
        "amount": 230,
        "account_id": "068d81de-be46-4d59-a20a-8d3c168504ff",
        "currency_code": "USD",
        "reference_id": "a0f60930-b58d-57b7-bb90-463c9f42e2f2",
        "metadata": {
            "LinkedAccountID": "068d81de-be46-4d59-a20a-8d3c168504ff",
            "LinkedTransactionID": "a0f60930-b58d-57b7-bb90-463c9f42e2f2",
            "OperationType": "Fee Transfer",
            "timestamp": "2024-08-16 17:08:28.431509 +0300 +03 m=+28.017843918"
        },
        "status": "SUCCESS",
        "transaction_type": "OUTGOING_FEE",
        "created_at": "2024-08-16T17:08:28.433178+03:00",
        "updated_at": "2024-08-16T17:08:28.445264+03:00"
    },
    "SourceTransaction": {
        "id": "24dabadb-c354-40d6-b16d-8eb2f9b0054f",
        "user_id": "3f1060e4-18f2-4b23-bc93-bd04ae29f9e7",
        "amount": 100,
        "account_id": "068d81de-be46-4d59-a20a-8d3c168504ff",
        "currency_code": "USD",
        "reference_id": "0d305986-3fde-5792-a867-a625d898e9fb",
        "metadata": {
            "DestinationAccountID": "60f4c37f-3509-41a8-b3c1-836c0bb70d39",
            "LinkedAccountID": "068d81de-be46-4d59-a20a-8d3c168504ff",
            "LinkedTransactionID": "0d305986-3fde-5792-a867-a625d898e9fb",
            "OperationType": "Transfer",
            "timestamp": "2024-08-16 17:08:28.404834 +0300 +03 m=+27.991168376"
        },
        "status": "SUCCESS",
        "transaction_type": "OUTBOUND",
        "created_at": "2024-08-16T17:08:28.413771+03:00",
        "updated_at": "2024-08-16T17:08:28.439805+03:00"
    },
    "DestinationTransaction": {
        "id": "a9bc31dd-e12c-4d28-b4c1-730ba9e1eae2",
        "user_id": "32f40751-b202-4598-bbd2-86d9fda9136a",
        "amount": 100,
        "account_id": "60f4c37f-3509-41a8-b3c1-836c0bb70d39",
        "currency_code": "USD",
        "reference_id": "beecbcc3-bd14-5b6a-93f3-b320694527c4",
        "metadata": {
            "DestinationAccountID": "60f4c37f-3509-41a8-b3c1-836c0bb70d39",
            "LinkedAccountID": "60f4c37f-3509-41a8-b3c1-836c0bb70d39",
            "LinkedTransactionID": "beecbcc3-bd14-5b6a-93f3-b320694527c4",
            "OperationType": "Transfer",
            "SourceAccountID": "068d81de-be46-4d59-a20a-8d3c168504ff",
            "timestamp": "2024-08-16 17:08:28.433887 +0300 +03 m=+28.020221084"
        },
        "status": "SUCCESS",
        "transaction_type": "INBOUND",
        "created_at": "2024-08-16T17:08:28.434807+03:00",
        "updated_at": "2024-08-16T17:08:28.443416+03:00"
    }
}
```
## Screenshot from Temporal UI Transfer workflow:

![Transfer Workflow](https://i.ibb.co/XVM6xJP/Screenshot-2024-08-18-at-17-04-05.png)


### Detailed view 
![Transfer Workflow Detail](https://i.ibb.co/QrdtC5n/Screenshot-2024-08-18-at-17-05-29.png)
