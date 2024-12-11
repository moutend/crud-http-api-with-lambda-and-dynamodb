package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type Item struct {
	ID    string  `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

const (
	tableName = "http-crud-tutorial-items"
)

var (
	dbClient *dynamodb.Client
)

func deleteItemById(ctx context.Context, event events.APIGatewayV2HTTPRequest) (*events.APIGatewayV2HTTPResponse, error) {
	_, err := dbClient.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(tableName),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{
				Value: event.PathParameters["id"],
			},
		},
	})
	if err != nil {
		return nil, err
	}

	body, err := json.Marshal(struct {
		ID string `json:"id"`
	}{
		ID: event.PathParameters["id"],
	})
	if err != nil {
		return nil, err
	}

	response := events.APIGatewayV2HTTPResponse{
		StatusCode: http.StatusAccepted,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: string(body),
	}

	return &response, nil
}

func putItemById(ctx context.Context, event events.APIGatewayV2HTTPRequest) (*events.APIGatewayV2HTTPResponse, error) {
	var requestBody *bytes.Buffer
	if event.IsBase64Encoded {
		decodedBytes, err := base64.StdEncoding.DecodeString(event.Body)
		if err != nil {
			return nil, err
		}

		requestBody = bytes.NewBuffer(decodedBytes)
	} else {
		requestBody = bytes.NewBufferString(event.Body)
	}

	var item Item
	if err := json.NewDecoder(requestBody).Decode(&item); err != nil {
		return nil, err
	}

	_, err := dbClient.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{
				Value: item.ID,
			},
			"name": &types.AttributeValueMemberS{
				Value: item.Name,
			},
			"price": &types.AttributeValueMemberN{
				Value: fmt.Sprintf("%f", item.Price),
			},
		},
	})
	if err != nil {
		return nil, err
	}

	body, err := json.Marshal(item)
	if err != nil {
		return nil, err
	}

	response := events.APIGatewayV2HTTPResponse{
		StatusCode: http.StatusCreated,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: string(body),
	}

	return &response, nil
}

func getItemById(ctx context.Context, event events.APIGatewayV2HTTPRequest) (*events.APIGatewayV2HTTPResponse, error) {
	output, err := dbClient.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{
				Value: event.PathParameters["id"],
			},
		},
	})
	if err != nil {
		return nil, err
	}
	if output.Item == nil {
		response := events.APIGatewayV2HTTPResponse{
			StatusCode: http.StatusNotFound,
		}
		return &response, nil
	}

	var item Item
	if err := attributevalue.UnmarshalMap(output.Item, &item); err != nil {
		return nil, err
	}

	body, err := json.Marshal(item)
	if err != nil {
		return nil, err
	}

	response := events.APIGatewayV2HTTPResponse{
		StatusCode: http.StatusOK,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: string(body),
	}

	return &response, nil
}

func getItems(ctx context.Context, event events.APIGatewayV2HTTPRequest) (*events.APIGatewayV2HTTPResponse, error) {
	output, err := dbClient.Scan(ctx, &dynamodb.ScanInput{
		TableName: aws.String(tableName),
	})
	if err != nil {
		return nil, err
	}

	var items []*Item
	if err := attributevalue.UnmarshalListOfMaps(output.Items, &items); err != nil {
		return nil, err
	}

	body, err := json.Marshal(items)
	if err != nil {
		return nil, err
	}

	response := events.APIGatewayV2HTTPResponse{
		StatusCode: http.StatusOK,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: string(body),
	}

	return &response, nil
}

func errorResponse(err error) (*events.APIGatewayV2HTTPResponse, error) {
	body, err := json.Marshal(struct {
		Error string `json:"error"`
	}{
		Error: fmt.Sprintf("%s", err),
	})
	if err != nil {
		return nil, err
	}

	response := events.APIGatewayV2HTTPResponse{
		StatusCode: http.StatusInternalServerError,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: string(body),
	}

	return &response, nil
}

func handler(ctx context.Context, event events.APIGatewayV2HTTPRequest) (*events.APIGatewayV2HTTPResponse, error) {
	var response *events.APIGatewayV2HTTPResponse
	var err error

	switch event.RouteKey {
	case "DELETE /items/{id}":
		response, err = deleteItemById(ctx, event)
	case "PUT /items":
		response, err = putItemById(ctx, event)
	case "GET /items/{id}":
		response, err = getItemById(ctx, event)
	case "GET /items":
		response, err = getItems(ctx, event)
	default:
		err = fmt.Errorf("Unsupported RouteKey: %q", event.RouteKey)
	}
	if err != nil {
		return errorResponse(err)
	}

	return response, nil
}

func init() {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		panic(err)
	}
	dbClient = dynamodb.NewFromConfig(cfg)
}

func main() {
	lambda.Start(handler)
}
