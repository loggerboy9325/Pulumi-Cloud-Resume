package dynamodb

import (
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/dynamodb"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func CreateDynamoDB(ctx *pulumi.Context) error {
	table, err := dynamodb.NewTable(ctx, "test", &dynamodb.TableArgs{
		Name:          pulumi.String("test-table"),
		ReadCapacity:  pulumi.Int(10),
		WriteCapacity: pulumi.Int(10),
		HashKey:       pulumi.String("exampleHashKey"),
		Attributes: dynamodb.TableAttributeArray{
			&dynamodb.TableAttributeArgs{
				Name: pulumi.String("exampleHashKey"),
				Type: pulumi.String("S"),
			},
		},
	})
	if err != nil {
		println(err)
	}
	ctx.Export("table", table)

	return nil
}
