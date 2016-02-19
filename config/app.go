package config

import (
	"fmt"
	"time"
)

// Config variables
const (
	ListenAddr           = "listen-addr"
	BuildDir             = "build-dir"
	ServerTimeout        = "server-timeout"
	URI                  = "vili-uri"
	StaticLiveReload     = "static-live-reload"
	Environments         = "environments"
	ProdEnvs             = "prod-envs"
	ApprovalEnvs         = "approval-envs"
	LogDebug             = "log-debug"
	LogJSON              = "log-json"
	RedisPort            = "redis-port"
	RedisDB              = "redis-db"
	OktaEntrypoint       = "okta-entrypoint"
	OktaIssuer           = "okta-issuer"
	OktaCert             = "okta-cert"
	OktaDomain           = "okta-domain"
	GithubToken          = "github-token"
	GithubOwner          = "github-owner"
	GithubRepo           = "github-repo"
	GithubContentsPath   = "github-contents-path"
	QuayToken            = "quay-token"
	QuayNamespace        = "quay-namespace"
	FirebaseURL          = "firebase-url"
	FirebaseSecret       = "firebase-secret"
	SlackToken           = "slack-token"
	SlackChannel         = "slack-channel"
	SlackUsername        = "slack-username"
	SlackEmoji           = "slack-emoji"
	SlackDeployUsernames = "slack-deploy-usernames"
)

// KubernetesURL returns the config variable name for robot tokens
func KubernetesURL(env string) string {
	return fmt.Sprintf("kube-%s-url", env)
}

// KubernetesNamespace returns the config variable name for robot tokens
func KubernetesNamespace(env string) string {
	return fmt.Sprintf("kube-%s-namespace", env)
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
	return Require(
		BuildDir,
		URI,
		Environments,
		RedisPort,
		GithubToken,
		GithubOwner,
		GithubRepo,
		GithubContentsPath,
		QuayToken,
		QuayNamespace,
		SlackToken,
		SlackChannel,
	)
}
