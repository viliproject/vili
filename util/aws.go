package util

import (
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
)

var awsIAMARNRegexp = regexp.MustCompile("arn:aws:iam::([0-9]{12}):(.*)")

// GetAWSAccountNumber returns the account number for the given aws config
func GetAWSAccountNumber(awsConfig *aws.Config) string {
	svc := iam.New(session.New(awsConfig))
	output, err := svc.GetUser(&iam.GetUserInput{})
	if err == nil {
		match := awsIAMARNRegexp.FindStringSubmatch(*output.User.Arn)
		if match != nil {
			return match[1]
		}
	}
	return ""
}
