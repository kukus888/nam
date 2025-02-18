#!/bin/bash

URL=http://localhost:8080

curl -L $URL/api/rest/applications -X POST -d '{"Name":"TestApp01","Port":36125,"Type":"Spring"}'

curl -L $URL/api/rest/servers -X POST -d '{"Alias":"TestServer01","Hostname":"TestServer01.domain"}'