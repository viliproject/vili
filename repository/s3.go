package repository

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

var s3Service *S3Service

// S3Config is the S3 service configuration
type S3Config struct {
	Region          string
	AccessKeyID     string
	SecretAccessKey string
}

// S3Service is an implementation of the bundle Service interface
// It fetches bundles from S3
type S3Service struct {
	config *S3Config
	s3     *s3.S3
}

func newS3(c *S3Config) *S3Service {
	awsConfig := aws.NewConfig()
	creds := credentials.NewChainCredentials([]credentials.Provider{
		&credentials.StaticProvider{
			Value: credentials.Value{
				AccessKeyID:     c.AccessKeyID,
				SecretAccessKey: c.SecretAccessKey,
			},
		},
		&credentials.EnvProvider{},
		&credentials.SharedCredentialsProvider{},
		&ec2rolecreds.EC2RoleProvider{Client: ec2metadata.New(session.New())},
	})
	awsConfig.WithCredentials(creds)
	awsConfig.WithRegion(c.Region)
	return &S3Service{
		config: c,
		s3:     s3.New(session.New(awsConfig)),
	}
}

// InitS3 initializes the docker registry service
func InitS3(c *S3Config) error {
	s3Service = newS3(c)
	return nil
}

// GetRepository implements the Service interface
func (s *S3Service) GetRepository(ctx context.Context, bucket, key string, branches []string) ([]*Image, error) {
	images, err := s.getImagesForBranches(ctx, bucket, key, branches)
	if err != nil {
		return nil, err
	}

	sortByLastModified(images)
	return images, nil
}

func (s *S3Service) getImagesForBranches(ctx context.Context, bucket, key string, branchNames []string) (images []*Image, err error) {
	prefix := strings.TrimSuffix(strings.TrimPrefix(key, "/"), "/") + "/"
	err = s.s3.ListObjectsV2Pages(&s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
		Prefix: aws.String(prefix),
	}, func(output *s3.ListObjectsV2Output, lastPage bool) bool {
		for _, obj := range output.Contents {
			tag := strings.TrimSuffix(filepath.Base(*obj.Key), ".zip")
			image := &Image{
				Tag:          tag,
				LastModified: *obj.LastModified,
			}
			sepIndex := strings.LastIndex(tag, "-")
			if sepIndex != -1 {
				branchComponent, shaComponent := tag[:sepIndex], tag[sepIndex+1:]
				image.Revision = shaComponent
				for _, branchName := range branchNames {
					if branchComponent == slugFromBranch(branchName) {
						image.Branch = branchName
						images = append(images, image)
					}
				}
			}
		}
		return true
	})
	return
}
