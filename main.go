package main

import (
	"cloudresume/dynamodb"
	"cloudresume/lambda"
	"cloudresume/s3"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		err := s3.CreateS3Bucket(ctx)
		if err != nil {
			return err
		}
		err = lambda.CreateLambda(ctx)
		if err != nil {
			return err
		}
		err = dynamodb.CreateDynamoDB(ctx)
		if err != nil {
			return err
		}
		return nil
	})
}
