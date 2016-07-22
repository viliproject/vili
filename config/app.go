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
	ProdEnvs                = "prod-envs"
	ApprovalEnvs            = "approval-envs"
	DefaultEnv              = "default-env"
	LogDebug                = "log-debug"
	LogJSON                 = "log-json"
	RedisPort               = "redis-port"
	RedisDB                 = "redis-db"
	OktaEntrypoint          = "okta-entrypoint"
	OktaIssuer              = "okta-issuer"
	OktaCert                = "okta-cert"
	OktaDomain              = "okta-domain"
	GithubToken             = "github-token"
	GithubOwner             = "github-owner"
	GithubRepo              = "github-repo"
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
)

// KubernetesURL returns the config variable name for robot tokens
func KubernetesURL(env string) string {
	return fmt.Sprintf("kube-%s-url", env)
}

// KubernetesNamespace returns the config variable name for robot tokens
func KubernetesNamespace(env string) string {
	return fmt.Sprintf("kube-%s-namespace", env)
}

// KubernetesClientCert returns the config variable name for robot tokens
func KubernetesClientCert(env string) string {
	return fmt.Sprintf("kube-%s-client-cert", env)
}

// KubernetesClientKey returns the config variable name for robot tokens
func KubernetesClientKey(env string) string {
	return fmt.Sprintf("kube-%s-client-key", env)
}

// GithubEnvContentsPath returns the config variable name for robot tokens
func GithubEnvContentsPath(env string) string {
	return fmt.Sprintf("github-envs-%s-contents-path", env)
}

// InitApp initializes the config
func InitApp() error {
	SetDefault(ListenAddr, ":80")
	SetDefault(ServerTimeout, time.Second*30)
	SetDefault(SlackUsername, "vili")
	SetDefault(ProdEnvs, "prod")
	SetDefault(ApprovalEnvs, "preprod")
	SetDefault(RegistryBranchDelimiter, "-")
	SetDefault(DockerMode, "registry")
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
