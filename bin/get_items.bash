#!/bin/bash

api_endpoint=$(terraform output -raw api_endpoint)

curl -i "${api_endpoint}/items"
