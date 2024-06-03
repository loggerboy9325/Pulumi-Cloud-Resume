package s3

import (
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/s3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func CreateS3Bucket(ctx *pulumi.Context) error {
	bucket, err := s3.NewBucket(ctx, "testpulumibucket", nil)
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
		})
		if err != nil {
			return err
		}
		ctx.Export("bucke", bucket.ID())
	}
	return nil
}
