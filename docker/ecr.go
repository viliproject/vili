package docker

import (
	"errors"
	"strconv"
	"strings"
	"sync"
	"time"

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
	BranchDelimiter string
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
	service = newECR(c)
	return nil
}

// GetRepository implements the Service interface
func (s *ECRService) GetRepository(repo string, withBranches bool) ([]*Image, error) {
	var waitGroup sync.WaitGroup
	imagesChan := make(chan getImagesResult, len(branches))

	if withBranches {
		for _, branch := range branches {
			if branch != "master" {
				waitGroup.Add(1)
				go func(branch string) {
					defer waitGroup.Done()
					images, err := s.getImagesForBranch(repo, branch)
					imagesChan <- getImagesResult{images: images, err: err}
				}(branch)
			}
		}
	}

	images, err := s.getImagesForBranch(repo, "master")
	if err != nil {
		return nil, err
	}

	waitGroup.Wait()
	close(imagesChan)
	for result := range imagesChan {
		if result.err != nil {
			return nil, result.err
		}
		images = append(images, result.images...)
	}

	sortByLastModified(images)
	return images, nil
}

// GetTag implements the Service interface
func (s *ECRService) GetTag(repo, branch, tag string) (string, error) {
	fullRepoName := s.getRepositoryForBranch(repo, branch)

	resp, err := s.ecr.BatchGetImage(&ecr.BatchGetImageInput{
		ImageIds: []*ecr.ImageIdentifier{
			{
				ImageTag: aws.String(tag),
			},
		},
		RepositoryName: aws.String(fullRepoName),
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
func (s *ECRService) FullName(repo, branch, tag string) (string, error) {
	resp, err := s.ecr.DescribeRepositories(&ecr.DescribeRepositoriesInput{
		RepositoryNames: []*string{
			aws.String(s.getRepositoryForBranch(repo, branch)),
		},
	})
	if err != nil {
		return "", err
	}
	if len(resp.Repositories) < 1 {
		return "", errors.New("Repository not found")
	}
	return *resp.Repositories[0].RepositoryUri + ":" + tag, nil
}

func (s *ECRService) getImagesForBranch(repoName, branchName string) ([]*Image, error) {
	fullRepoName := s.getRepositoryForBranch(repoName, branchName)

	var images []*Image
	var nextToken *string

	for {
		resp, err := s.ecr.ListImages(&ecr.ListImagesInput{
			RepositoryName: aws.String(fullRepoName),
			NextToken:      nextToken,
		})
		if err != nil {
			return images, err
		}
		for _, imageID := range resp.ImageIds {
			if imageID.ImageTag != nil {
				image := &Image{
					Tag:    *imageID.ImageTag,
					Branch: branchName,
				}
				sepIndex := strings.LastIndex(*imageID.ImageTag, "-")
				if sepIndex != -1 {
					dateComponent, shaComponent := (*imageID.ImageTag)[:sepIndex], (*imageID.ImageTag)[sepIndex+1:]
					unixSecs, err := strconv.ParseInt(dateComponent, 10, 0)
					if err != nil {
						continue
					}
					image.Revision = shaComponent
					image.LastModified = time.Unix(unixSecs, 0)
				}
				images = append(images, image)
			}
		}
		if resp.NextToken == nil {
			return images, nil
		}
		nextToken = resp.NextToken
	}
}

func (s *ECRService) getRepositoryForBranch(repoName, branchName string) string {
	if s.config.Namespace != "" {
		repoName = s.config.Namespace + "/" + repoName
	}
	if branchName == "master" {
		return repoName
	}
	return repoName + s.config.BranchDelimiter + strings.ToLower(branchName)
}
