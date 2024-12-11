#!/bin/bash

api_endpoint=$(terraform output -raw api_endpoint)
data='{"id":"ABC","name":"Apple","price":12.34}'

curl -i -X PUT --data ${data} "${api_endpoint}/items"
