package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"net/http"
)

type APIGatewayProxyResponse struct {
	StatusCode int               `json:"statusCode"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body"`
}

func handler(req *http.Request) (APIGatewayProxyResponse, error) {
	// Your logic here
	responseBody := "{\"message\": \"Success!\"}"

	response := APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: responseBody,
	}

	return response, nil
}

func main() {
	lambda.Start(handler)
}
