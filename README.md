# Overview

This repository contains a Go implementation of the following AWS tutorial:

> **Tutorial:** Create a CRUD HTTP API with Lambda and DynamoDB  
> [https://docs.aws.amazon.com/apigateway/latest/developerguide/http-api-dynamo-db.html](https://docs.aws.amazon.com/apigateway/latest/developerguide/http-api-dynamo-db.html)

## Deployment

To deploy the application, run:

```bash
./bin/deploy.bash
```

## Testing

To test the application, run:

```bash
./bin/get_items.bash   # Fetch items from the database
./bin/put_item.bash    # Add a new item to the database
./bin/get_items.bash   # Verify the added item
```
## Cleanup

To remove the deployed resources, run:

```bash
terraform destroy -auto-approve
```
