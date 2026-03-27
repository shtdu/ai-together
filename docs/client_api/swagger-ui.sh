#!/bin/bash
# 
# This script is used to start the swagger-ui server.
# It is used to serve the swagger-ui files to the client.
# the default port is 8082.
# proxy the api to by pass cors policy.
echo "access the swagger-ui at http://localhost:8082/swagger/"
npx http-server --cors --proxy http://localhost:8081
