package vili

import (
	"encoding/json"
	"fmt"
	"net/http"

	"gopkg.in/labstack/echo.v1"

	"github.com/airware/vili/config"
	"github.com/airware/vili/environments"
	"github.com/airware/vili/firebase"
	"github.com/airware/vili/public"
	"github.com/airware/vili/session"
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
	URI        string                     `json:"uri"`
	User       *session.User              `json:"user"`
	DefaultEnv string                     `json:"defaultEnv"`
	Envs       []environments.Environment `json:"envs"`
	Firebase   FirebaseConfig             `json:"firebase"`
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
			fmt.Sprintf(homeTemplate, "null", fmt.Sprintf("/static/app-%s.js", public.GetHash(staticLiveReload))),
		)
	}
	return appHandler(c)
}

func appHandler(c *echo.Context) error {
	envs := environments.Environments()
	user, _ := c.Get("user").(*session.User)

	firebaseToken, err := firebase.NewToken(user)
	if err != nil {
		return err
	}

	appConfig := AppConfig{
		URI:        config.GetString(config.URI),
		User:       user,
		DefaultEnv: config.GetString(config.DefaultEnv),
		Envs:       envs,
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
		fmt.Sprintf(homeTemplate, appConfigBytes, fmt.Sprintf("/static/app-%s.js", public.GetHash(staticLiveReload))),
	)
}
