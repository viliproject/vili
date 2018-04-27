package repository

import (
	"errors"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
)

// ECRConfig is the ECR service configuration
type ECRConfig struct {
	Region          string
	AccessKeyID     string
	SecretAccessKey string
	Namespace       string
	RegistryID      *string
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
	dockerService = newECR(c)
	return nil
}

// GetRepository implements the Service interface
func (s *ECRService) GetRepository(repo string, branches []string) ([]*Image, error) {
	images, err := s.getImagesForBranches(repo, branches)
	if err != nil {
		return nil, err
	}

	sortByLastModified(images)
	return images, nil
}

// GetTag implements the Service interface
func (s *ECRService) GetTag(repo, tag string) (string, error) {
	fullRepoName := s.fullRepositoryName(repo)

	resp, err := s.ecr.BatchGetImage(&ecr.BatchGetImageInput{
		ImageIds: []*ecr.ImageIdentifier{
			{
				ImageTag: aws.String(tag),
			},
		},
		RepositoryName: aws.String(fullRepoName),
		RegistryId:     s.config.RegistryID,
	})
	if err != nil {
		return "", err
	}

	if len(resp.Failures) > 0 {
		return "", errors.New(*resp.Failures[0].FailureReason)
	}

	return *resp.Images[0].ImageId.ImageDigest, nil
}

// FullName implements the Service interface
func (s *ECRService) FullName(repo, tag string) (string, error) {
	resp, err := s.ecr.DescribeRepositories(&ecr.DescribeRepositoriesInput{
		RepositoryNames: []*string{
			aws.String(s.fullRepositoryName(repo)),
		},
		RegistryId: s.config.RegistryID,
	})
	if err != nil {
		return "", err
	}
	if len(resp.Repositories) < 1 {
		return "", errors.New("Repository not found")
	}
	return *resp.Repositories[0].RepositoryUri + ":" + tag, nil
}

func (s *ECRService) getImagesForBranches(repoName string, branchNames []string) ([]*Image, error) {
	fullRepoName := s.fullRepositoryName(repoName)

	var images []*Image
	var nextToken *string

	for {
		resp, err := s.ecr.DescribeImages(&ecr.DescribeImagesInput{
			RepositoryName: &fullRepoName,
			NextToken:      nextToken,
			RegistryId:     s.config.RegistryID,
			Filter: &ecr.DescribeImagesFilter{
				TagStatus: aws.String("TAGGED"),
			},
		})
		if err != nil {
			return images, err
		}
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
		if resp.NextToken == nil {
			return images, nil
		}
		nextToken = resp.NextToken
	}
}

func (s *ECRService) fullRepositoryName(repoName string) string {
	if s.config.Namespace != "" {
		return s.config.Namespace + "/" + repoName
	}
	return repoName
}
