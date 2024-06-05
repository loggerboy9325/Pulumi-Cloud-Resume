package main

import (
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"golang.org/x/net/context"
)

var tableName = "Pulumi-Resume-Table"

func handler(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Get the value from the request body
	valueToWrite := event.Body

	// Create a new DynamoDB session
	sess := session.Must(session.NewSession())
	svc := dynamodb.New(sess)

	// Put the value into DynamoDB table
	_, err := svc.PutItem(&dynamodb.PutItemInput{
		Item: map[string]*dynamodb.AttributeValue{
			"Value": {
				S: aws.String(valueToWrite),
			},
		},
		TableName: aws.String(tableName),
	})
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       fmt.Sprintf("Failed to write value to DynamoDB: %v", err),
		}, nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       "Value written to DynamoDB table successfully.",
	}, nil
}

func main() {
	lambda.Start(handler)
}
