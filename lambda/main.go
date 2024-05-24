package main

import (
	"context"

	"github.com/aws/aws-lambda-go/lambda"
)

// Define the request and response types
type MyEvent struct {
	Name string `json:"name"`
}

type MyResponse struct {
	Message string `json:"message"`
}

func handleRequest(ctx context.Context, event MyEvent) (MyResponse, error) {
	response := MyResponse{
		Message: "Hello, World!",
	}
	return response, nil
}

func main() {
	lambda.Start(handleRequest)
}
