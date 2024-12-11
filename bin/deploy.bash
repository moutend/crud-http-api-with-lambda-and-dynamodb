#!/bin/bash

set -ex

terraform apply -target=aws_ecr_repository.http_crud_tutorial -auto-approve
repo_url=$(terraform output -raw ecr_repository_url)
repo_host=${repo_url%%/*}
docker build --platform linux/arm64 -t my-golang-example:v1 .
docker tag my-golang-example:v1 "${repo_url}:v1"
docker logout ${repo_host}
aws.bash ecr get-login-password --profile admin --region us-east-1 \
  | docker login --username AWS --password-stdin ${repo_host}
docker push "${repo_url}:v1"
terraform apply -auto-approve
