package repository

import (
	"context"
	"errors"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
)

var ecrService *ECRService

// ECRConfig is the ECR service configuration
type ECRConfig struct {
	Region          string
	AccessKeyID     string
	SecretAccessKey string
}

// ECRService is an implementation of the docker Service interface
// It fetches docker images
type ECRService struct {
	config *ECRConfig
	ecr    *ecr.ECR
}

func newECR(c *ECRConfig) *ECRService {
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
	return &ECRService{
		config: c,
		ecr:    ecr.New(session.New(awsConfig)),
	}
}

// InitECR initializes the docker registry service
func InitECR(c *ECRConfig) error {
	ecrService = newECR(c)
	return nil
}

// GetRepository implements the Service interface
func (s *ECRService) GetRepository(ctx context.Context, accountID, repoName string, branches []string) ([]*Image, error) {
	images, err := s.getImagesForBranches(ctx, accountID, repoName, branches)
	if err != nil {
		return nil, err
	}

	sortByLastModified(images)
	return images, nil
}

// GetTag implements the Service interface
func (s *ECRService) GetTag(ctx context.Context, accountID, repoName, tag string) (string, error) {
	resp, err := s.ecr.BatchGetImageWithContext(ctx, &ecr.BatchGetImageInput{
		ImageIds: []*ecr.ImageIdentifier{
			{
				ImageTag: aws.String(tag),
			},
		},
		RegistryId:     aws.String(accountID),
		RepositoryName: aws.String(repoName),
	})
	if err != nil {
		return "", err
	}

	if len(resp.Failures) > 0 {
		return "", errors.New(*resp.Failures[0].FailureReason)
	}

	return *resp.Images[0].ImageId.ImageDigest, nil
}

func (s *ECRService) getImagesForBranches(ctx context.Context, accountID, repoName string, branchNames []string) (images []*Image, err error) {
	err = s.ecr.DescribeImagesPagesWithContext(ctx, &ecr.DescribeImagesInput{
		RegistryId:     aws.String(accountID),
		RepositoryName: aws.String(repoName),
		Filter: &ecr.DescribeImagesFilter{
			TagStatus: aws.String("TAGGED"),
		},
	}, func(resp *ecr.DescribeImagesOutput, lastPage bool) bool {
		for _, imageDetails := range resp.ImageDetails {
			for _, tag := range imageDetails.ImageTags {
				image := &Image{
					Tag:          *tag,
					LastModified: *imageDetails.ImagePushedAt,
				}
				sepIndex := strings.LastIndex(*tag, "-")
				if sepIndex != -1 {
					branchComponent, shaComponent := (*tag)[:sepIndex], (*tag)[sepIndex+1:]
					image.Revision = shaComponent
					for _, branchName := range branchNames {
						if branchComponent == slugFromBranch(branchName) {
							image.Branch = branchName
							images = append(images, image)
						}
					}
				}
			}
		}
		return true
	})
	return
}
