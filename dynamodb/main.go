package dynamodb

import (
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/dynamodb"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func CreateDynamoDB(ctx *pulumi.Context) error {
	table, err := dynamodb.NewTable(ctx, "Pulumi-Resume-Table", &dynamodb.TableArgs{
		Name:        pulumi.String("visitors-1"),
		BillingMode: pulumi.String("PAY_PER_REQUEST"),
		HashKey:     pulumi.String("ID"),
		Attributes: dynamodb.TableAttributeArray{
			&dynamodb.TableAttributeArgs{
				Name: pulumi.String("ID"),
				Type: pulumi.String("S"),
			},
		},
	})
	if err != nil {
		println(err)
	}
	ctx.Export("table", table.Arn)

	return nil
}
