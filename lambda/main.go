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

	function, err := lambda.NewFunction(ctx, "Pulumi-Resume", &lambda.FunctionArgs{
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
	api, err := apigateway.NewRestApi(ctx, "myApi", &apigateway.RestApiArgs{
		Description: pulumi.String("My api gateway"),
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
	Method, err := apigateway.NewMethod(ctx, "mymethod", &apigateway.MethodArgs{
		RestApi:       api.ID(),
		ResourceId:    resource.ID(),
		HttpMethod:    pulumi.String("GET"),
		Authorization: pulumi.String("NONE"),
	})
	if err != nil {
		return nil
	}
	integ, err := apigateway.NewIntegration(ctx, "myIntegration", &apigateway.IntegrationArgs{
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
	_, err = apigateway.NewDeployment(ctx, "Mydeployment", &apigateway.DeploymentArgs{
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
