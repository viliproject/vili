package vili

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/labstack/echo"

	"github.com/airware/vili/config"
	"github.com/airware/vili/errors"
	"github.com/airware/vili/firebase"
	"github.com/airware/vili/log"
	"github.com/airware/vili/public"
	"github.com/airware/vili/session"
	"github.com/airware/vili/templates"
)

const homeTemplate = `
<!doctype html>
<html class="full-height">
    <head>
        <title>Vili</title>
        <link rel='shortcut icon' type='image/x-icon' href='https://static.airware.com/app/favicon.ico' />
        <link rel="stylesheet" href="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.5/css/bootstrap.min.css" />
        <script type="application/javascript">window.appconfig = %s;</script>
        <script type="application/javascript" src="%s"></script>
    </head>
    <body></body>
</html>`

// AppConfig is the frontend configuration
type AppConfig struct {
	URI          string              `json:"uri"`
	User         *session.User       `json:"user"`
	Envs         []string            `json:"envs"`
	ProdEnvs     []string            `json:"prodEnvs"`
	ApprovalEnvs []string            `json:"approvalEnvs"`
	EnvApps      map[string][]string `json:"envApps"`
	EnvJobs      map[string][]string `json:"envJobs"`
	Firebase     FirebaseConfig      `json:"firebase"`
}

// FirebaseConfig is the Firebase configuration
type FirebaseConfig struct {
	URL   string `json:"url"`
	Token string `json:"token"`
}

func homeHandler(c *echo.Context) error {
	if c.Get("user") == nil {
		staticLiveReload := config.GetBool(config.StaticLiveReload)
		return c.HTML(
			http.StatusOK,
			homeTemplate,
			"null",
			fmt.Sprintf("/static/app-%s.js", public.GetHash(staticLiveReload)),
		)
	}
	return appHandler(c)
}

func appHandler(c *echo.Context) error {
	envs := config.GetStringSlice(config.Environments)
	envApps := make(map[string][]string)
	envJobs := make(map[string][]string)

	failed := false
	var wg sync.WaitGroup
	var appsMutex sync.Mutex
	var jobsMutex sync.Mutex
	funcs := []func(env string){
		func(env string) {
			defer wg.Done()
			controllers, err := templates.Controllers(env)
			if err != nil {
				log.Error(err)
				failed = true
			}
			appsMutex.Lock()
			envApps[env] = controllers
			appsMutex.Unlock()
		},
		func(env string) {
			defer wg.Done()
			pods, err := templates.Pods(env)
			if err != nil {
				log.Error(err)
				failed = true
			}
			jobsMutex.Lock()
			envJobs[env] = pods
			jobsMutex.Unlock()
		},
	}

	wg.Add(len(funcs) * len(envs))
	for _, f := range funcs {
		for _, env := range envs {
			go f(env)
		}
	}
	wg.Wait()
	if failed {
		return errors.New("failed github call")
	}

	user, _ := c.Get("user").(*session.User)

	firebaseToken, err := firebase.NewToken(user)
	if err != nil {
		return err
	}

	appConfig := AppConfig{
		URI:          config.GetString(config.URI),
		User:         user,
		Envs:         envs,
		ProdEnvs:     config.GetStringSlice(config.ProdEnvs),
		ApprovalEnvs: config.GetStringSlice(config.ApprovalEnvs),
		EnvApps:      envApps,
		EnvJobs:      envJobs,
		Firebase: FirebaseConfig{
			URL:   config.GetString(config.FirebaseURL),
			Token: firebaseToken,
		},
	}
	appConfigBytes, err := json.Marshal(appConfig)
	if err != nil {
		return err
	}
	staticLiveReload := config.GetBool(config.StaticLiveReload)
	return c.HTML(
		http.StatusOK,
		homeTemplate,
		string(appConfigBytes),
		fmt.Sprintf("/static/app-%s.js", public.GetHash(staticLiveReload)),
	)
}
