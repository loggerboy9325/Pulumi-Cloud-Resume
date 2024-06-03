package lambda

import (
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/iam"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/lambda"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func CreateLambda(ctx *pulumi.Context) error {
	role, err := iam.NewRole(ctx, "lambdaRole", &iam.RoleArgs{
		AssumeRolePolicy: pulumi.String(`
		{
		 "Version": "2012-10-17",
		 "Statement": [
		   {
		     "Action": "sts:AssumeRole",
		     "Principal": {
		       "Service": "lambda.amazonaws.com"
		     },
		     "Effect": "Allow",
		     "Sid": ""
		   }
		 ]
		}
		`),
	})
	if err != nil {
		return err
	}

	// Attach necessary policies to the role
	_, err = iam.NewRolePolicyAttachment(ctx, "lambdaRoleAttachment", &iam.RolePolicyAttachmentArgs{
		Role:      role.Name,
		PolicyArn: pulumi.String("arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"),
	})
	if err != nil {
		return err
	}

	function, err := lambda.NewFunction(ctx, "myLambdaFunction", &lambda.FunctionArgs{
		Code:        pulumi.NewFileArchive("lambda/function.zip"),
		Runtime:     pulumi.String("provided.al2023"),
		Role:        role.Arn,
		Handler:     pulumi.String("main"),
		Description: pulumi.String("My Lambda function"),
		MemorySize:  pulumi.Int(128),
		Timeout:     pulumi.Int(5),
		Environment: &lambda.FunctionEnvironmentArgs{
			Variables: pulumi.StringMap{
				"ENV_VAR": pulumi.String("value"),
			},
		},
	})
	if err != nil {
		return err
	}
	ctx.Export("functionname", function.Name)
	return nil
}
