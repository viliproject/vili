package config

import (
	"fmt"
	"time"
)

// Config variables
const (
	ListenAddr              = "listen-addr"
	BuildDir                = "build-dir"
	ServerTimeout           = "server-timeout"
	URI                     = "vili-uri"
	StaticLiveReload        = "static-live-reload"
	AppCert                 = "app-cert"
	AppPrivateKey           = "app-private-key"
	Environments            = "environments"
	ApprovalProdEnvs        = "approval-prod-envs"
	IgnoredEnvs             = "ignored-envs"
	DefaultEnv              = "default-env"
	EnvKubernetesNamespaces = "env-kube-namespaces"
	LogDebug                = "log-debug"
	LogJSON                 = "log-json"
	RedisPort               = "redis-port"
	RedisDB                 = "redis-db"
	AuthService             = "auth-service"
	SAMLMetadataURL         = "saml-metadata-url"
	HardcodedTokenUsers     = "hardcoded-token-users"
	GithubToken             = "github-token"
	GithubOwner             = "github-owner"
	GithubRepo              = "github-repo"
	GithubDefaultBranch     = "github-default-branch"
	GithubContentsPath      = "github-contents-path"
	DockerMode              = "docker-mode"
	BundleMode              = "bundle-mode"
	FunctionsMode           = "functions-mode"
	AWSRegion               = "aws-region"
	AWSRepositoryBucket     = "aws-repository-bucket"
	AWSAccessKeyID          = "aws-access-key-id"
	AWSSecretAccessKey      = "aws-secret-access-key"
	RegistryURL             = "registry-url"
	RegistryBranchDelimiter = "registry-branch-delimiter"
	RegistryNamespace       = "registry-namespace"
	RegistryUsername        = "registry-username"
	RegistryPassword        = "registry-password"
	BundleNamespace         = "bundle-namespace"
	ECRAccountID            = "ecr-account-id"
	FirebaseURL             = "firebase-url"
	FirebaseSecret          = "firebase-secret"
	SlackToken              = "slack-token"
	SlackChannel            = "slack-channel"
	SlackUsername           = "slack-username"
	SlackEmoji              = "slack-emoji"
	SlackDeployUsernames    = "slack-deploy-usernames"
	RolloutTimeout          = "rollout-timeout"
	JobRunTimeout           = "job-run-timeout"
	CIProvider              = "ci-provider"
	CircleCIToken           = "circleci-token"
)

// EnvRepositoryBranches returns the config variable name for the
// repository branches for the given env
func EnvRepositoryBranches(env string) string {
	return fmt.Sprintf("env-%s-repository-branches", env)
}

// KubeConfigPath returns the config variable name for the kubeconfig path
func KubeConfigPath(env string) string {
	return fmt.Sprintf("%s-kubeconfig-path", env)
}

// GithubEnvContentsPath returns the config variable name for the contents path for
// a given environment
func GithubEnvContentsPath(env string) string {
	return fmt.Sprintf("github-envs-%s-contents-path", env)
}

// InitApp initializes the config
func InitApp() error {
	SetDefault(ListenAddr, ":80")
	SetDefault(ServerTimeout, time.Second*30)
	SetDefault(SlackUsername, "vili")
	SetDefault(ApprovalProdEnvs, "preprod prod")
	SetDefault(RegistryBranchDelimiter, "-")
	SetDefault(DockerMode, "registry")
	SetDefault(RolloutTimeout, 10*time.Minute)
	SetDefault(JobRunTimeout, 10*time.Minute)
	return Require(
		BuildDir,
		URI,
		Environments,
		RedisPort,
		GithubToken,
		GithubOwner,
		GithubRepo,
		GithubContentsPath,
	)
}
