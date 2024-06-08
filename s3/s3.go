package s3

import (
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/cloudfront"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/s3"
	synced "github.com/pulumi/pulumi-synced-folder/sdk/go/synced-folder"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func CreateS3Bucket(ctx *pulumi.Context) error {

	// Create an S3 bucket
	bucket, err := s3.NewBucket(ctx, "pulumi-test-bucket", &s3.BucketArgs{
		Website: &s3.BucketWebsiteArgs{
			IndexDocument: pulumi.String("index.html"),
			ErrorDocument: pulumi.String("error.html"),
		},
	})
	if err != nil {
		return err
	}

	oai, err := cloudfront.NewOriginAccessIdentity(ctx, "oai", &cloudfront.OriginAccessIdentityArgs{
		Comment: pulumi.String("OAI for S3 bucket"),
	})
	if err != nil {
		return err
	}

	publicAccessBlock, err := s3.NewBucketPublicAccessBlock(ctx, "public-access-block", &s3.BucketPublicAccessBlockArgs{
		Bucket:          bucket.Bucket,
		BlockPublicAcls: pulumi.Bool(false),
	})
	if err != nil {
		return err
	}

	// Put an object into the S3 bucket
	_, err = synced.NewS3BucketFolder(ctx, "synced-folder", &synced.S3BucketFolderArgs{
		Path:       pulumi.String("./images/"),
		BucketName: bucket.Bucket,
		Acl:        pulumi.String("private"),
	}, pulumi.DependsOn([]pulumi.Resource{publicAccessBlock}))
	if err != nil {
		return err
	}

	_, err = s3.NewBucketObject(ctx, "index.html", &s3.BucketObjectArgs{
		Bucket: bucket.Bucket,
		Source: pulumi.NewFileAsset("index.html"),
	})
	if err != nil {
		return err
	}

	_, err = s3.NewBucketObject(ctx, "script.js", &s3.BucketObjectArgs{
		Bucket: bucket.Bucket,
		Source: pulumi.NewFileAsset("script.js"),
	})
	if err != nil {
		return err
	}

	_, err = s3.NewBucketObject(ctx, "error.html", &s3.BucketObjectArgs{
		Bucket: bucket.Bucket,
		Source: pulumi.NewFileAsset("error.html"),
	})
	if err != nil {
		return err
	}

	policyJSON := pulumi.All(bucket.Arn, oai.IamArn).ApplyT(func(args []interface{}) string {
		bucketArn := args[0].(string)
		iamArn := args[1].(string)
		return `{
				"Version": "2012-10-17",
				"Statement": [
					{
						"Effect": "Allow",
						"Principal": {
							"AWS": "` + iamArn + `"
						},
						"Action": "s3:GetObject",
						"Resource": "` + bucketArn + `/*"
					}
				]
			}`
	}).(pulumi.StringOutput)

	_, err = s3.NewBucketPolicy(ctx, "bucketPolicy", &s3.BucketPolicyArgs{
		Bucket: bucket.ID(),
		Policy: policyJSON,
	})
	if err != nil {
		return err
	}

	cdn, err := cloudfront.NewDistribution(ctx, "cdn", &cloudfront.DistributionArgs{
		Enabled: pulumi.Bool(true),
		Origins: cloudfront.DistributionOriginArray{
			&cloudfront.DistributionOriginArgs{
				OriginId:   bucket.Arn,
				DomainName: bucket.BucketRegionalDomainName,
				S3OriginConfig: &cloudfront.DistributionOriginS3OriginConfigArgs{
					OriginAccessIdentity: oai.CloudfrontAccessIdentityPath,
				},
			},
		},
		DefaultCacheBehavior: &cloudfront.DistributionDefaultCacheBehaviorArgs{
			TargetOriginId:       bucket.Arn,
			ViewerProtocolPolicy: pulumi.String("redirect-to-https"),
			AllowedMethods: pulumi.StringArray{
				pulumi.String("GET"),
				pulumi.String("HEAD"),
				pulumi.String("OPTIONS"),
			},
			CachedMethods: pulumi.StringArray{
				pulumi.String("GET"),
				pulumi.String("HEAD"),
				pulumi.String("OPTIONS"),
			},
			DefaultTtl: pulumi.Int(600),
			MaxTtl:     pulumi.Int(600),
			MinTtl:     pulumi.Int(600),
			ForwardedValues: &cloudfront.DistributionDefaultCacheBehaviorForwardedValuesArgs{
				QueryString: pulumi.Bool(false),
				Cookies: &cloudfront.DistributionDefaultCacheBehaviorForwardedValuesCookiesArgs{
					Forward: pulumi.String("none"),
				},
			},
		},
		PriceClass: pulumi.String("PriceClass_200"),

		Restrictions: &cloudfront.DistributionRestrictionsArgs{
			GeoRestriction: &cloudfront.DistributionRestrictionsGeoRestrictionArgs{
				RestrictionType: pulumi.String("none"),
			},
		},
		ViewerCertificate: &cloudfront.DistributionViewerCertificateArgs{
			CloudfrontDefaultCertificate: pulumi.Bool(true),
		},
	})
	if err != nil {
		return err
	}

	// Export the name of the bucket and the CloudFront distribution domain name
	ctx.Export("bucketName", bucket.Bucket)
	ctx.Export("cloudFrontDomainName", bucket.BucketRegionalDomainName)
	ctx.Export("originURL", pulumi.Sprintf("http://%v", bucket.WebsiteEndpoint))
	ctx.Export("originHostname", bucket.WebsiteEndpoint)
	ctx.Export("cdnURL", pulumi.Sprintf("https://%v", cdn.DomainName))
	ctx.Export("cdnHostname", cdn.DomainName)

	return nil
}
