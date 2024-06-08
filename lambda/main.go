package lambda

import (
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/apigateway"
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

	_, err = iam.NewRolePolicy(ctx, "lambda-Policy", &iam.RolePolicyArgs{
		Role: role.Name,
		Policy: pulumi.String(`{
				"Version": "2012-10-17",
				"Statement": [{
					"Effect": "Allow",
					"Action": [
						"dynamodb:GetItem",
						"dynamodb:PutItem",
						"dynamodb:UpdateItem",
						"dynamodb:DeleteItem"
					],
					"Resource": "arn:aws:dynamodb:us-east-1:936791179343:table/visitors-1"
				}]
			}`),
	}, pulumi.DependsOn([]pulumi.Resource{role}))
	if err != nil {
		return err
	}

	function, err := lambda.NewFunction(ctx, "Pulumi-Resume", &lambda.FunctionArgs{
		Code:        pulumi.NewFileArchive("lambda/lambda-handler/lambda-handler.zip"),
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
	api, err := apigateway.NewRestApi(ctx, "Pulumi-Reseume-Api", &apigateway.RestApiArgs{
		Description: pulumi.String("My api gateway"),
	})
	if err != nil {
		return err
	}

	_, err = lambda.NewPermission(ctx, "apiGatewayPermission", &lambda.PermissionArgs{
		Action:    pulumi.String("lambda:InvokeFunction"),
		Function:  function.Name,
		Principal: pulumi.String("apigateway.amazonaws.com"),
		SourceArn: pulumi.Sprintf("%v/*/*/*", api.ExecutionArn),
	})
	if err != nil {
		return err
	}

	resource, err := apigateway.NewResource(ctx, "Resource", &apigateway.ResourceArgs{
		RestApi:  api.ID(),
		PathPart: pulumi.String("my-resource"),
		ParentId: api.RootResourceId,
	})
	if err != nil {
		return err
	}
	Method, err := apigateway.NewMethod(ctx, "Pulumi-Resume-method", &apigateway.MethodArgs{
		RestApi:       api.ID(),
		ResourceId:    resource.ID(),
		HttpMethod:    pulumi.String("GET"),
		Authorization: pulumi.String("NONE"),
	})
	if err != nil {
		return nil
	}
	integ, err := apigateway.NewIntegration(ctx, "Pulumi-Resume-Integration", &apigateway.IntegrationArgs{
		RestApi:               api.ID(),
		ResourceId:            resource.ID(),
		HttpMethod:            pulumi.String("GET"),
		IntegrationHttpMethod: pulumi.String("POST"),
		Type:                  pulumi.String("AWS_PROXY"),
		Uri:                   function.InvokeArn,
	})
	if err != nil {
		return err
	}
	_, err = apigateway.NewDeployment(ctx, "Pulumi-deployment", &apigateway.DeploymentArgs{
		RestApi:   api.ID(),
		StageName: pulumi.String("prod"),
	}, pulumi.DependsOn([]pulumi.Resource{integ, Method}))
	if err != nil {
		return err
	}
	ctx.Export("functionname", function.Name)
	ctx.Export("apiURL", function.InvokeArn)
	ctx.Export("url", pulumi.Sprintf("https://%s.execute-api.%s.amazonaws.com/prod/myresource", api.ID(), "us-east-1"))
	return nil
}
