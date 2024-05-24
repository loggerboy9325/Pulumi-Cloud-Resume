package main

import (
	"fmt"

	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/iam"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/lambda"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/s3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		// Create an AWS resource (S3 Bucket)
		bucket, err := s3.NewBucket(ctx, "Cloud-Resume", &s3.BucketArgs{
			Website: &s3.BucketWebsiteArgs{
				IndexDocument: pulumi.String("index.html"),
			},
		})
		if err != nil {
			return err
		}

		ownershipControls, err := s3.NewBucketOwnershipControls(ctx, "ownership-controls", &s3.BucketOwnershipControlsArgs{
			Bucket: bucket.ID(),
			Rule: &s3.BucketOwnershipControlsRuleArgs{
				ObjectOwnership: pulumi.String("ObjectWriter"),
			},
		})
		if err != nil {
			return err
		}

		publicAccessBlock, err := s3.NewBucketPublicAccessBlock(ctx, "public-access-block", &s3.BucketPublicAccessBlockArgs{
			Bucket:                bucket.ID(),
			BlockPublicAcls:       pulumi.Bool(true),
			IgnorePublicAcls:      pulumi.Bool(true),
			BlockPublicPolicy:     pulumi.Bool(true),
			RestrictPublicBuckets: pulumi.Bool(true),
		})
		if err != nil {
			return err
		}

		files := []string{"index.html", "script.js", "images/CCP.png", "images/portrait.jpg",
			"images/SAA.png", "images/SAP.png", "ResumeFinal.pdf"}

		for _, file := range files {
			_, err := s3.NewBucketObject(ctx, file, &s3.BucketObjectArgs{
				Bucket:      bucket.ID(),
				Source:      pulumi.NewFileAsset(file),
				ContentType: pulumi.String("text/html"),
				Acl:         pulumi.String("public-read"),
			}, pulumi.DependsOn([]pulumi.Resource{
				publicAccessBlock,
				ownershipControls,
			}))
			if err != nil {
				return err
			}
		}

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

		// Create a Lambda function
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

		// Export the name of the bucket
		ctx.Export("bucketEndpoint", bucket.WebsiteEndpoint.ApplyT(func(websiteEndpoint string) (string, error) {
			return fmt.Sprintf("http://%v", websiteEndpoint), nil
		}).(pulumi.StringOutput))

		ctx.Export("lambdaFunctionName", function.Name)

		return nil
	})
}
