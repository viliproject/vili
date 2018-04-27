package functions

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/airware/vili/environments"
	"github.com/airware/vili/log"
	"github.com/airware/vili/repository"
	"github.com/airware/vili/templates"
	"github.com/airware/vili/util"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
)

const lambdaActiveAliasName = "active"

// LambdaConfig is the Lambda service configuration
type LambdaConfig struct {
	Region          string
	AccessKeyID     string
	SecretAccessKey string
}

// LambdaService is an implementation of the docker Service interface
// It fetches docker images
type LambdaService struct {
	config        *LambdaConfig
	lambda        *lambda.Lambda
	accountNumber string
}

func newLambda(c *LambdaConfig) *LambdaService {
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

	return &LambdaService{
		config:        c,
		lambda:        lambda.New(session.New(awsConfig)),
		accountNumber: util.GetAWSAccountNumber(awsConfig),
	}
}

// InitLambda initializes the docker registry service
func InitLambda(c *LambdaConfig) error {
	service = newLambda(c)
	return nil
}

// List implements the Service interface
func (s *LambdaService) List(ctx context.Context, env string) (functions []Function, err error) {
	environment, err := environments.Get(env)
	if err != nil {
		return
	}
	functionsChan := make(chan (Function))

	go func() {
		var wg sync.WaitGroup
		for _, funcName := range environment.Functions {
			wg.Add(1)
			go func(funcName string) {
				defer wg.Done()
				function, err := s.Get(ctx, env, funcName)
				if err != nil {
					log.WithError(err).Errorf("failed getting function %s - %s", env, funcName)
					return
				}
				if function != nil {
					functionsChan <- function
				}
			}(funcName)
		}

		wg.Wait()
		close(functionsChan)
	}()

	for function := range functionsChan {
		functions = append(functions, function)
	}
	return
}

// Get implements the Service interface
func (s *LambdaService) Get(ctx context.Context, env, name string) (function Function, err error) {
	var config *lambda.FunctionConfiguration
	var versions []*LambdaFunctionVersion
	var activeAliasConfig *lambda.AliasConfiguration
	var wg sync.WaitGroup

	functionName := makeFunctionName(env, name)
	// get function
	wg.Add(1)
	go func() {
		defer wg.Done()
		output, err := s.lambda.GetFunctionWithContext(
			ctx,
			&lambda.GetFunctionInput{
				FunctionName: aws.String(functionName),
			},
		)
		if err != nil {
			if !isResourceNotFound(err) {
				log.WithError(err).Errorf("failed getting function - %T - %s", err, functionName)
			}
			return
		}
		config = output.Configuration
	}()
	// get function versions
	wg.Add(1)
	go func() {
		defer wg.Done()
		listVersionsInput := &lambda.ListVersionsByFunctionInput{
			FunctionName: aws.String(functionName),
		}
		for {
			output, err := s.lambda.ListVersionsByFunctionWithContext(ctx, listVersionsInput)
			if err != nil {
				if !isResourceNotFound(err) {
					log.WithError(err).Errorf("failed getting function versions - %T - %s", err, functionName)
				}
				return
			}
			for _, v := range output.Versions {
				if *v.Version == "$LATEST" {
					continue
				}
				versions = append(versions, makeLambdaFunctionVersion(v))
			}
			if output.NextMarker == nil {
				break
			}
			listVersionsInput.Marker = output.NextMarker
		}
	}()
	// get function active alias
	wg.Add(1)
	go func() {
		defer wg.Done()
		output, err := s.lambda.GetAliasWithContext(
			ctx,
			&lambda.GetAliasInput{
				FunctionName: aws.String(functionName),
				Name:         aws.String(lambdaActiveAliasName),
			},
		)
		if err != nil {
			if !isResourceNotFound(err) {
				log.WithError(err).Errorf("failed getting function alias - %T - %s", err, functionName)
			}
			return
		}
		activeAliasConfig = output
	}()
	wg.Wait()

	if config != nil {
		function = makeLambdaFunction(env, name, config, versions, activeAliasConfig)
	}
	return
}

// Deploy implements the Service interface
func (s *LambdaService) Deploy(ctx context.Context, env, name string, spec *FunctionDeploySpec) (err error) {
	functionName := makeFunctionName(env, name)

	// get the bucket and key from bundle name
	bundleName, err := repository.BundleFullName(name, spec.Tag)
	if err != nil {
		return
	}
	bundleURL, err := url.Parse(bundleName)
	if err != nil {
		return
	}

	updateFunctionCode := &lambda.UpdateFunctionCodeInput{
		FunctionName: aws.String(functionName),
		S3Bucket:     aws.String(bundleURL.Host),
		S3Key:        aws.String(strings.TrimPrefix(bundleURL.Path, "/")),
	}

	// get the spec and populate it
	functionTemplate, err := templates.Function(env, spec.Branch, name)
	if err != nil {
		return
	}
	functionTemplate, err = functionTemplate.Populate(map[string]string{
		"Namespace":       env,
		"AWSAcountNumber": s.accountNumber,
	})
	if err != nil {
		return
	}

	functionSpec := new(LambdaFunctionSpec)
	err = functionTemplate.Parse(functionSpec)
	if err != nil {
		return
	}

	function, err := s.Get(ctx, env, name)
	if err != nil {
		return
	}

	description, err := json.Marshal(&LambdaFunctionVersion{
		Env:        env,
		Tag:        spec.Tag,
		Branch:     spec.Branch,
		DeployedBy: spec.DeployedBy,
	})
	if err != nil {
		return
	}

	tags := map[string]*string{
		"vili/namespace": aws.String(env),
	}

	updateFunctionConfig := &lambda.UpdateFunctionConfigurationInput{
		FunctionName: aws.String(functionName),
		Description:  aws.String(string(description)),
		Runtime:      aws.String(functionSpec.Runtime),
		Handler:      aws.String(functionSpec.Handler),
		Role:         aws.String(functionSpec.Role),
		Environment:  getLambdaEnvironment(functionSpec.Environment),
	}
	if functionSpec.MemorySize > 0 {
		updateFunctionConfig.MemorySize = aws.Int64(functionSpec.MemorySize)
	}
	if functionSpec.Timeout > 0 {
		updateFunctionConfig.Timeout = aws.Int64(functionSpec.Timeout)
	}
	if functionSpec.TracingConfig != nil {
		updateFunctionConfig.TracingConfig = &lambda.TracingConfig{
			Mode: aws.String(functionSpec.TracingConfig.Mode),
		}
	}
	if functionSpec.VPCConfig != nil {
		updateFunctionConfig.VpcConfig = &lambda.VpcConfig{
			SecurityGroupIds: aws.StringSlice(functionSpec.VPCConfig.SecurityGroupIDs),
			SubnetIds:        aws.StringSlice(functionSpec.VPCConfig.SubnetIDs),
		}
	}

	var funcConfig *lambda.FunctionConfiguration
	aliasExists := false
	if function == nil {
		// create function
		funcConfig, err = s.lambda.CreateFunctionWithContext(ctx, &lambda.CreateFunctionInput{
			FunctionName: aws.String(functionName),
			Tags:         tags,
			Code: &lambda.FunctionCode{
				S3Bucket: updateFunctionCode.S3Bucket,
				S3Key:    updateFunctionCode.S3Key,
			},
			Publish:       aws.Bool(true),
			Description:   updateFunctionConfig.Description,
			Environment:   updateFunctionConfig.Environment,
			Handler:       updateFunctionConfig.Handler,
			MemorySize:    updateFunctionConfig.MemorySize,
			Role:          updateFunctionConfig.Role,
			Runtime:       updateFunctionConfig.Runtime,
			Timeout:       updateFunctionConfig.Timeout,
			TracingConfig: updateFunctionConfig.TracingConfig,
			VpcConfig:     updateFunctionConfig.VpcConfig,
		})
		if err != nil {
			return
		}
	} else {
		lambdaFunction := function.(*LambdaFunction)

		updateFunctionCode.Publish = aws.Bool(true)
		// update configuration
		_, err = s.lambda.UpdateFunctionConfigurationWithContext(ctx, updateFunctionConfig)
		if err != nil {
			return
		}
		// update tags
		_, err = s.lambda.TagResourceWithContext(ctx, &lambda.TagResourceInput{
			Resource: lambdaFunction.config.FunctionArn,
			Tags:     tags,
		})
		if err != nil {
			return
		}
		// update code and publish new version
		funcConfig, err = s.lambda.UpdateFunctionCodeWithContext(ctx, updateFunctionCode)
		if err != nil {
			return
		}

		aliasExists = lambdaFunction.activeAliasConfig != nil
	}

	if aliasExists {
		// set the active alias to the new version
		_, err = s.lambda.UpdateAliasWithContext(ctx, &lambda.UpdateAliasInput{
			FunctionName:    aws.String(functionName),
			FunctionVersion: aws.String(*funcConfig.Version),
			Name:            aws.String(lambdaActiveAliasName),
		})
	} else {
		// set the active alias to the new version
		_, err = s.lambda.CreateAliasWithContext(ctx, &lambda.CreateAliasInput{
			FunctionName:    aws.String(functionName),
			FunctionVersion: aws.String(*funcConfig.Version),
			Name:            aws.String(lambdaActiveAliasName),
		})
	}
	return
}

// Rollback implements the Service interface
func (s *LambdaService) Rollback(ctx context.Context, env, name, version string) (err error) {
	_, err = s.lambda.UpdateAliasWithContext(ctx, &lambda.UpdateAliasInput{
		FunctionName:    aws.String(makeFunctionName(env, name)),
		FunctionVersion: aws.String(version),
		Name:            aws.String(lambdaActiveAliasName),
	})
	return
}

func makeLambdaFunction(env, name string, config *lambda.FunctionConfiguration, versions []*LambdaFunctionVersion, activeAliasConfig *lambda.AliasConfiguration) *LambdaFunction {
	// TODO sort versions by time
	sort.SliceStable(versions, func(i, j int) bool {
		vi := versions[i]
		vj := versions[j]
		return *vi.config.LastModified > *vj.config.LastModified
	})
	f := &LambdaFunction{
		env:               env,
		name:              name,
		config:            config,
		activeAliasConfig: activeAliasConfig,
		versions:          versions,
	}
	return f
}

func makeLambdaFunctionVersion(config *lambda.FunctionConfiguration) *LambdaFunctionVersion {
	v := &LambdaFunctionVersion{
		config: config,
	}
	// try to unmarshal description into the function, but ignore any errors
	json.Unmarshal([]byte(*config.Description), v)
	return v
}

// LambdaFunction is a wrapper for a lambda function configuration
type LambdaFunction struct {
	env               string
	name              string
	config            *lambda.FunctionConfiguration
	activeAliasConfig *lambda.AliasConfiguration
	versions          []*LambdaFunctionVersion
}

// LambdaFunctionVersion is a wrapper for a lambda function version
type LambdaFunctionVersion struct {
	config *lambda.FunctionConfiguration

	Env        string `json:"env"`
	Tag        string `json:"tag"`
	Branch     string `json:"branch"`
	DeployedBy string `json:"deployedBy"`
}

// GetName implements the Function interface
func (f *LambdaFunction) GetName() string {
	return f.name
}

// GetEnv implements the Function interface
func (f *LambdaFunction) GetEnv() string {
	activeVersion := f.GetActiveVersion()
	if activeVersion != nil {
		return activeVersion.GetEnv()
	}
	return ""
}

// GetActiveVersion implements the Function interface
func (f *LambdaFunction) GetActiveVersion() FunctionVersion {
	for _, version := range f.versions {
		if *f.activeAliasConfig.FunctionVersion == version.GetVersion() {
			return version
		}
	}
	return nil
}

// GetVersions implements the Function interface
func (f *LambdaFunction) GetVersions() (versions []FunctionVersion) {
	for _, v := range f.versions {
		versions = append(versions, v)
	}
	return
}

// GetVersion implements the FunctionVersion interface
func (v *LambdaFunctionVersion) GetVersion() string {
	return *v.config.Version
}

// GetLastModified implements the FunctionVersion interface
func (v *LambdaFunctionVersion) GetLastModified() time.Time {
	t, _ := time.Parse("2006-01-02T15:04:05.000-0700", *v.config.LastModified)
	return t
}

// GetEnv implements the FunctionVersion interface
func (v *LambdaFunctionVersion) GetEnv() string {
	return v.Env
}

// GetTag implements the FunctionVersion interface
func (v *LambdaFunctionVersion) GetTag() string {
	return v.Tag
}

// GetBranch implements the FunctionVersion interface
func (v *LambdaFunctionVersion) GetBranch() string {
	return v.Branch
}

// GetDeployedBy implements the FunctionVersion interface
func (v *LambdaFunctionVersion) GetDeployedBy() string {
	return v.DeployedBy
}

func makeFunctionName(env, name string) string {
	return fmt.Sprintf("%s__%s", env, name)
}

// LambdaFunctionSpec is the spec for the lambda function configuration
type LambdaFunctionSpec struct {
	Runtime       string                           `json:"runtime"`
	Handler       string                           `json:"handler"`
	Role          string                           `json:"role"`
	MemorySize    int64                            `json:"memorySize"`
	Timeout       int64                            `json:"timeout"`
	Environment   map[string]string                `json:"environment"`
	TracingConfig *LambdaFunctionTracingConfigSpec `json:"tracingConfig,omitempty"`
	VPCConfig     *LambdaFunctionVPCConfigSpec     `json:"vpcConfig,omitempty"`
}

// LambdaFunctionTracingConfigSpec is the spec for the lambda function tracing config
type LambdaFunctionTracingConfigSpec struct {
	Mode string `json:"mode"`
}

// LambdaFunctionVPCConfigSpec is the spec for the lambda function vpc config
type LambdaFunctionVPCConfigSpec struct {
	SecurityGroupIDs []string `json:"securityGroupIds"`
	SubnetIDs        []string `json:"subnetIds"`
}

func getLambdaEnvironment(environment map[string]string) *lambda.Environment {
	variables := map[string]*string{}
	for k, v := range environment {
		variables[k] = aws.String(v)
	}
	return &lambda.Environment{
		Variables: variables,
	}
}

func isResourceNotFound(err error) bool {
	if awsErr, ok := err.(awserr.Error); ok {
		return awsErr.Code() == "ResourceNotFoundException"
	}
	return false
}
