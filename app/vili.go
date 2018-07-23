package vili

import (
	"net/url"
	"sync"

	"github.com/airware/vili/api"
	"github.com/airware/vili/auth"
	"github.com/airware/vili/config"
	"github.com/airware/vili/environments"
	"github.com/airware/vili/firebase"
	"github.com/airware/vili/functions"
	"github.com/airware/vili/git"
	"github.com/airware/vili/kube"
	"github.com/airware/vili/log"
	"github.com/airware/vili/middleware"
	"github.com/airware/vili/public"
	"github.com/airware/vili/redis"
	"github.com/airware/vili/repository"
	"github.com/airware/vili/server"
	"github.com/airware/vili/session"
	"github.com/airware/vili/slack"
	"github.com/airware/vili/stats"
	"github.com/airware/vili/templates"
	"github.com/airware/vili/util"
	"github.com/labstack/echo"
)

const appName = "vili"

// App is the wrapper for the tyr app
type App struct {
	server *server.Server
}

// New returns a new server instance
func New() *App {
	// check config
	if err := config.Init(); err != nil {
		log.Error(err)
		return nil
	}

	// set up logging
	log.Init(&log.Config{
		LogJSON:  config.GetBool(config.LogJSON),
		LogDebug: config.GetBool(config.LogDebug),
	})

	stats.TrackMemStats()
	environments.Init()

	// Set everything up in parallel
	var wg sync.WaitGroup
	initFunctions := []func(){
		// set up static assets
		func() {
			defer wg.Done()
			err := public.LoadStats(config.GetString(config.BuildDir))
			if err != nil {
				log.Fatal(err)
			}
		},

		// set up the kubernetes client
		func() {
			defer wg.Done()
			envConfigs := make(map[string]*kube.EnvConfig)
			envKubeNamespaces := config.GetStringSliceMap(config.EnvKubernetesNamespaces)
			for _, env := range environments.Environments() {
				envConfigs[env.Name] = &kube.EnvConfig{
					Namespace:      envKubeNamespaces[env.Name],
					KubeConfigPath: config.GetString(config.KubeConfigPath(env.Name)),
				}
			}
			err := kube.Init(&kube.Config{
				EnvConfigs:            envConfigs,
				DefaultKubeConfigPath: config.GetString(config.KubeConfigPath(config.GetString(config.DefaultEnv))),
			})
			if err != nil {
				log.Fatal(err)
			}
		},

		// set up the firebase client
		func() {
			defer wg.Done()
			err := firebase.Init(&firebase.Config{
				URL:    config.GetString(config.FirebaseURL),
				Secret: config.GetString(config.FirebaseSecret),
			})
			if err != nil {
				log.Fatal(err)
			}
		},

		// set up the redis client
		func() {
			defer wg.Done()
			urlp, err := url.Parse(config.GetString(config.RedisPort))
			if err != nil {
				log.Fatal(err)
			}
			err = redis.Init(&redis.Config{
				Addr: urlp.Host,
				DB:   config.GetInt(config.RedisDB),
			})
			if err != nil {
				log.Fatal(err)
			}
		},

		// set up the git service
		func() {
			defer wg.Done()
			git.InitGithub(&git.GithubConfig{
				Token:         config.GetString(config.GithubToken),
				Owner:         config.GetString(config.GithubOwner),
				Repo:          config.GetString(config.GithubRepo),
				DefaultBranch: config.GetString(config.GithubDefaultBranch),
			})
		},

		// set up the templates service
		func() {
			defer wg.Done()
			envContentsPaths := make(map[string]string, 0)
			for _, env := range environments.Environments() {
				envContentsPath := config.GetString(config.GithubEnvContentsPath(env.Name))
				if envContentsPath == "" {
					envContentsPath = config.GetString(config.GithubContentsPath)
				}
				envContentsPaths[env.Name] = envContentsPath
			}
			envContentsPaths[config.GetString(config.DefaultEnv)] = config.GetString(config.GithubContentsPath)
			templates.InitGit(&templates.GitConfig{
				EnvContentsPaths: envContentsPaths,
			})
		},

		// set up the docker registry
		func() {
			defer wg.Done()
			err := repository.InitRegistry(&repository.RegistryConfig{
				Username: config.GetString(config.RegistryUsername),
				Password: config.GetString(config.RegistryPassword),
			})
			if err != nil {
				log.Fatal(err)
			}
		},

		// set up the ECR registry
		func() {
			defer wg.Done()
			if config.IsSet(config.AWSRegion) && config.IsSet(config.AWSAccessKeyID) && config.IsSet(config.AWSSecretAccessKey) {
				err := repository.InitECR(&repository.ECRConfig{
					Region:          config.GetString(config.AWSRegion),
					AccessKeyID:     config.GetString(config.AWSAccessKeyID),
					SecretAccessKey: config.GetString(config.AWSSecretAccessKey),
				})
				if err != nil {
					log.Fatal(err)
				}
			}
		},

		// set up the S3 repository
		func() {
			defer wg.Done()
			if config.IsSet(config.AWSRegion) && config.IsSet(config.AWSAccessKeyID) && config.IsSet(config.AWSSecretAccessKey) {
				err := repository.InitS3(&repository.S3Config{
					Region:          config.GetString(config.AWSRegion),
					AccessKeyID:     config.GetString(config.AWSAccessKeyID),
					SecretAccessKey: config.GetString(config.AWSSecretAccessKey),
				})
				if err != nil {
					log.Fatal(err)
				}
			}
		},

		// set up functions
		func() {
			defer wg.Done()
			switch config.GetString(config.FunctionsMode) {
			case "lambda":
				err := functions.InitLambda(&functions.LambdaConfig{
					Region:          config.GetString(config.AWSRegion),
					AccessKeyID:     config.GetString(config.AWSAccessKeyID),
					SecretAccessKey: config.GetString(config.AWSSecretAccessKey),
				})
				if err != nil {
					log.Fatal(err)
				}
			default:
				// functions support is not required
			}
		},

		// set up the session services
		func() {
			defer wg.Done()
			session.InitHardcodedService(&session.HardcodedConfig{
				TokenUsers: config.GetStringSliceMap(config.HardcodedTokenUsers),
			})
			session.InitRedisService(&session.RedisConfig{
				Secure: false,
			})
		},

		// set up the auth service
		func() {
			defer wg.Done()
			switch config.GetString(config.AuthService) {
			case "saml":
				err := auth.InitSAMLAuthService(&auth.SAMLConfig{
					URL:            config.GetString(config.URI),
					IDPMetadataURL: config.GetString(config.SAMLMetadataURL),
					SPCert:         config.GetString(config.AppCert),
					SPPrivateKey:   config.GetString(config.AppPrivateKey),
				})
				if err != nil {
					log.Fatal(err)
				}
			case "null":
				err := auth.InitNullAuthService()
				if err != nil {
					log.Fatal(err)
				}
			default:
				log.Fatalf("Unknown auth service %s", config.GetString(config.AuthService))
			}
		},
		// set up the slack service
		func() {
			defer wg.Done()
			slack.Init(&slack.Config{
				Token:           config.GetString(config.SlackToken),
				Channel:         config.GetString(config.SlackChannel),
				Username:        config.GetString(config.SlackUsername),
				Emoji:           config.GetString(config.SlackEmoji),
				DeployUsernames: util.NewStringSet(config.GetStringSlice(config.SlackDeployUsernames)),
			})
		},
	}
	wg.Add(len(initFunctions))
	for _, f := range initFunctions {
		go f()
	}
	wg.Wait()

	// Setup and start the webserver
	s := server.New(&server.Config{
		Name:         appName,
		Addr:         config.GetString(config.ListenAddr),
		Timeout:      config.GetDuration(config.ServerTimeout),
		HealthCheck:  healthCheck,
		ShutdownFunc: shutdown,
		Middleware: []echo.MiddlewareFunc{
			middleware.Session(),
		},
	})

	auth.AddHandlers(s)
	api.AddHandlers(s)
	s.Echo().GET("/static/:name", public.StaticHandler)
	s.Echo().GET("/", homeHandler)
	s.Echo().GET("/*", middleware.RequireUser(appHandler))
	return &App{
		server: s,
	}
}

// Start starts the app
func (a *App) Start() {
	go runDeployBot()
	go environments.WatchEnvs()
	a.server.Start()
}

// StartTest starts the test app
func (a *App) StartTest() string {
	return a.server.StartTest()
}

// StopTest stops the test server
func (a *App) StopTest() {
	a.server.StopTest()
}

func healthCheck() error {
	return nil
}

func shutdown() {
	auth.Cleanup()
	log.Info("waiting for deployments and slack bot")
	close(environments.ExitingChan)
	close(api.ExitingChan)
	close(kube.ExitingChan)
	close(slack.ExitingChan)
	close(firebase.ExitingChan)
	api.WaitGroup.Wait()
	slack.WaitGroup.Wait()
	log.Info("finished with deployments and slack bot")
}
