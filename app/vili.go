package vili

import (
	"net/url"
	"sync"
	"time"

	"github.com/airware/vili/api"
	"github.com/airware/vili/auth"
	"github.com/airware/vili/config"
	"github.com/airware/vili/docker"
	"github.com/airware/vili/environments"
	"github.com/airware/vili/firebase"
	"github.com/airware/vili/git"
	"github.com/airware/vili/kube"
	"github.com/airware/vili/log"
	"github.com/airware/vili/middleware"
	"github.com/airware/vili/public"
	"github.com/airware/vili/redis"
	"github.com/airware/vili/server"
	"github.com/airware/vili/session"
	"github.com/airware/vili/slack"
	"github.com/airware/vili/stats"
	"github.com/airware/vili/templates"
	"github.com/airware/vili/util"
	echo "gopkg.in/labstack/echo.v1"
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
				log.Panic(err)
			}
		},

		// set up the kubernetes client
		func() {
			defer wg.Done()
			envConfigs := make(map[string]*kube.EnvConfig)
			for _, env := range environments.Environments() {
				envConfigs[env.Name] = &kube.EnvConfig{
					URL:          config.GetString(config.KubernetesURL(env.Name)),
					Namespace:    config.GetString(config.KubernetesNamespace(env.Name)),
					ClientCert:   config.GetString(config.KubernetesClientCert(env.Name)),
					ClientCACert: config.GetString(config.KubernetesClientCACert(env.Name)),
					ClientKey:    config.GetString(config.KubernetesClientKey(env.Name)),
				}
			}
			err := kube.Init(&kube.Config{
				EnvConfigs: envConfigs,
			})
			if err != nil {
				log.Panic(err)
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
				log.Panic(err)
			}
		},

		// set up the redis client
		func() {
			defer wg.Done()
			urlp, err := url.Parse(config.GetString(config.RedisPort))
			if err != nil {
				log.Panic(err)
			}
			err = redis.Init(&redis.Config{
				Addr: urlp.Host,
				DB:   config.GetInt(config.RedisDB),
			})
			if err != nil {
				log.Panic(err)
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

		// set up the docker service
		func() {
			defer wg.Done()
			switch config.GetString(config.DockerMode) {
			case "registry":
				err := docker.InitRegistry(&docker.RegistryConfig{
					BaseURL:         config.GetString(config.RegistryURL),
					Username:        config.GetString(config.RegistryUsername),
					Password:        config.GetString(config.RegistryPassword),
					Namespace:       config.GetString(config.RegistryNamespace),
					BranchDelimiter: config.GetString(config.RegistryBranchDelimiter),
				})
				if err != nil {
					log.Panic(err)
				}
			case "ecr":
				ecrAccountID := config.GetString(config.ECRAccountID)
				var registryID *string
				if ecrAccountID != "" {
					registryID = &ecrAccountID
				}
				err := docker.InitECR(&docker.ECRConfig{
					Region:          config.GetString(config.AWSRegion),
					AccessKeyID:     config.GetString(config.AWSAccessKeyID),
					SecretAccessKey: config.GetString(config.AWSSecretAccessKey),
					Namespace:       config.GetString(config.RegistryNamespace),
					BranchDelimiter: config.GetString(config.RegistryBranchDelimiter),
					RegistryID:      registryID,
				})
				if err != nil {
					log.Panic(err)
				}
			default:
				log.Panic("invalid docker mode provided")
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
			case "okta":
				err := auth.InitOktaAuthService(&auth.OktaConfig{
					Entrypoint: config.GetString(config.OktaEntrypoint),
					Issuer:     config.GetString(config.OktaIssuer),
					Cert:       config.GetString(config.OktaCert),
					Domain:     config.GetString(config.OktaDomain),
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
	s.Echo().Get("/static/:name", public.StaticHandler)
	s.Echo().Get("/", homeHandler)
	s.Echo().Get("/*", middleware.RequireUser(appHandler))
	return &App{
		server: s,
	}
}

// Start starts the app
func (a *App) Start() {
	go a.monitorEnvs()
	go runDeployBot()
	a.server.Start()
}

func (a *App) monitorEnvs() {
	for {
		if err := environments.RefreshEnvs(); err != nil {
			log.Warn("Unable to detect environments: ", err)
		}
		time.Sleep(5 * time.Minute)
	}
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
	close(api.ExitingChan)
	close(kube.ExitingChan)
	close(slack.ExitingChan)
	close(firebase.ExitingChan)
	api.WaitGroup.Wait()
	slack.WaitGroup.Wait()
	log.Info("finished with deployments and slack bot")
}
