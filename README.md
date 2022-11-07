# Go Album API

## Introduction

This project is a showcase of Kafka (with Transactions) + JWT Auth + Rest API (GIN) + MongoDb 

It consist of two services, first is `AlbumAPI` which is responsible for JWT auth/login, Album creation/update/deletion/querying from MongoDb, sending to `AlbumProcessor` via Kafka which is intended to process received Album somehow and return it back to AlbumAPI service to update the final state to `Executed`


## Prerequisites

* Go 1.19
* Docker

## Run
1. Deploy local MongoDb + Kafka
```bash
cd albumapi
./startRs.sh # script must setup repica set for mongo db(3 nodes) + express (mongo UI) on port 10001; deploy local kafka with 3 brokers and create two topics for albums
```

2. Run AlbumAPI service
```bash

cd albumapi
go run .

```

3. Run AlbumProcessor service
```bash
cd albumprocessor
go run .
```
4. Then open Postman tests (AlbumJWT collection): https://www.getpostman.com/collections/b82f58f19887c4825e92
5. Execute the following sequence: `Register Alex`-> `Login Alex` -> Copy jwt token from `accessToken` response field -> Execute `Alex new album` with access token -> See in logs processing by two services -> Query `Alex all albums` -> See that price is doubled and status = `PROCESSING_FINISHED`