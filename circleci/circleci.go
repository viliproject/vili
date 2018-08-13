package circleci

import (
	"github.com/airware/vili/log"
	"github.com/jszwedko/go-circleci"
)

var config *Config
var client *circleci.Client

// Config is the circle configuration
type Config struct {
	Token   string
	BaseURL string
}

// Init Initializes the circle ci client
func Init(c *Config) error {
	config = c
	if c.Token == "" || c.BaseURL == "" {
		log.Warn("Missing circle ci Token")
	}
	client = &circleci.Client{Token: c.Token}
	return nil
}

// CircleBuild runs a build on circle ci for the defined branch
func CircleBuild(account, repo, branch string, buildParameters map[string]string) (*circleci.Build, error) {
	build, err := client.ParameterizedBuild(account, repo, branch, buildParameters)
	if err != nil {
		return nil, err
	}
	return build, nil
}
