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
	OktaEntrypoint          = "okta-entrypoint"
	OktaIssuer              = "okta-issuer"
	OktaCert                = "okta-cert"
	OktaDomain              = "okta-domain"
	HardcodedTokenUsers     = "hardcoded-token-users"
	GithubToken             = "github-token"
	GithubOwner             = "github-owner"
	GithubRepo              = "github-repo"
	GithubDefaultBranch     = "github-default-branch"
	GithubContentsPath      = "github-contents-path"
	RegistryURL             = "registry-url"
	RegistryBranchDelimiter = "registry-branch-delimiter"
	RegistryNamespace       = "registry-namespace"
	RegistryUsername        = "registry-username"
	RegistryPassword        = "registry-password"
	AWSRegion               = "aws-region"
	AWSAccessKeyID          = "aws-access-key-id"
	AWSSecretAccessKey      = "aws-secret-access-key"
	ECRAccountID            = "ecr-account-id"
	DockerMode              = "docker-mode"
	FirebaseURL             = "firebase-url"
	FirebaseSecret          = "firebase-secret"
	SlackToken              = "slack-token"
	SlackChannel            = "slack-channel"
	SlackUsername           = "slack-username"
	SlackEmoji              = "slack-emoji"
	SlackDeployUsernames    = "slack-deploy-usernames"
	RolloutTimeout          = "rollout-timeout"
	JobRunTimeout           = "job-run-timeout"
)

// EnvRepositoryBranches returns the config variable name for the
// repository branches for the given env
func EnvRepositoryBranches(env string) string {
	return fmt.Sprintf("env-%s-repository-branches", env)
}

// KubernetesURL returns the config variable name for the kube url
func KubernetesURL(env string) string {
	return fmt.Sprintf("kube-%s-url", env)
}

// KubernetesClientCert returns the config variable name for kube certs
func KubernetesClientCert(env string) string {
	return fmt.Sprintf("kube-%s-client-cert", env)
}

// KubernetesClientCACert returns the config variable name for kube CA certs
func KubernetesClientCACert(env string) string {
	return fmt.Sprintf("kube-%s-client-ca-cert", env)
}

// KubernetesClientKey returns the config variable name for kube private keys
func KubernetesClientKey(env string) string {
	return fmt.Sprintf("kube-%s-client-key", env)
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
		SlackToken,
		SlackChannel,
	)
}
