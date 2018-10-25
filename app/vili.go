package vili

import (
	"net/url"
	"sync"

	"github.com/labstack/echo"
	"github.com/viliproject/vili/api"
	"github.com/viliproject/vili/auth"
	"github.com/viliproject/vili/config"
	"github.com/viliproject/vili/environments"
	"github.com/viliproject/vili/firebase"
	"github.com/viliproject/vili/functions"
	"github.com/viliproject/vili/git"
	"github.com/viliproject/vili/kube"
	"github.com/viliproject/vili/log"
	"github.com/viliproject/vili/middleware"
	"github.com/viliproject/vili/public"
	"github.com/viliproject/vili/redis"
	"github.com/viliproject/vili/repository"
	"github.com/viliproject/vili/server"
	"github.com/viliproject/vili/session"
	"github.com/viliproject/vili/slack"
	"github.com/viliproject/vili/stats"
	"github.com/viliproject/vili/templates"
	"github.com/viliproject/vili/util"
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

		// set up the docker repository
		func() {
			defer wg.Done()
			switch config.GetString(config.DockerMode) {
			case "registry":
				err := repository.InitRegistry(&repository.RegistryConfig{
					BaseURL:   config.GetString(config.RegistryURL),
					Username:  config.GetString(config.RegistryUsername),
					Password:  config.GetString(config.RegistryPassword),
					Namespace: config.GetString(config.RegistryNamespace),
				})
				if err != nil {
					log.Fatal(err)
				}
			case "ecr":
				ecrAccountID := config.GetString(config.ECRAccountID)
				var registryID *string
				if ecrAccountID != "" {
					registryID = &ecrAccountID
				}
				err := repository.InitECR(&repository.ECRConfig{
					Region:          config.GetString(config.AWSRegion),
					AccessKeyID:     config.GetString(config.AWSAccessKeyID),
					SecretAccessKey: config.GetString(config.AWSSecretAccessKey),
					Namespace:       config.GetString(config.RegistryNamespace),
					RegistryID:      registryID,
				})
				if err != nil {
					log.Fatal(err)
				}
			default:
				log.Fatal("invalid docker mode provided")
			}
		},

		// set up the bundle repository
		func() {
			defer wg.Done()
			switch config.GetString(config.BundleMode) {
			case "s3":
				err := repository.InitS3(&repository.S3Config{
					Region:          config.GetString(config.AWSRegion),
					Bucket:          config.GetString(config.AWSRepositoryBucket),
					Namespace:       config.GetString(config.BundleNamespace),
					AccessKeyID:     config.GetString(config.AWSAccessKeyID),
					SecretAccessKey: config.GetString(config.AWSSecretAccessKey),
				})
				if err != nil {
					log.Fatal(err)
				}
			default:
				// bundle repository is not required
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
			case "basic":
				err := auth.InitBasicAuthService(&auth.BasicConfig{
					Users: config.GetStringSlice(config.BasicAuthUsers),
				})
				if err != nil {
					log.Fatal(err)
				}
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
			if config.IsSet(config.SlackToken) {
				slack.Init(&slack.Config{
					Token:           config.GetString(config.SlackToken),
					Channel:         config.GetString(config.SlackChannel),
					Username:        config.GetString(config.SlackUsername),
					Emoji:           config.GetString(config.SlackEmoji),
					DeployUsernames: util.NewStringSet(config.GetStringSlice(config.SlackDeployUsernames)),
				})
			}
		},
		// set up the ci client
		func() {
			defer wg.Done()
			err := api.InitializeCiClient(config.GetString(config.CIProvider))
			if err != nil {
				log.Fatal(err)
			}
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
