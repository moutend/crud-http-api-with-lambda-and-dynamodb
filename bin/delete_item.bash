#!/bin/bash

api_endpoint=$(terraform output -raw api_endpoint)
item_id="ABC"

curl -i -X DELETE "${api_endpoint}/items/${item_id}"
