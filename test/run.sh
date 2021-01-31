#!/bin/bash

docker run --name postgres -e POSTGRES_PASSWORD=password -e POSTGRES_USER=viktor -e POSTGRES_DB=test -d postgres

#analyze POSTGRES_HOST_AUTH_METHOD FOR TEST PURPOSE