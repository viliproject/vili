package vili

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/labstack/echo"

	"github.com/viliproject/vili/config"
	"github.com/viliproject/vili/environments"
	"github.com/viliproject/vili/public"
	"github.com/viliproject/vili/session"
)

const homeTemplate = `
<!doctype html>
<html class="full-height">
    <head>
        <title>Vili</title>
        <link rel='shortcut icon' type='image/x-icon' href='https://static.airware.com/app/favicon.ico' />
        <link rel="stylesheet" href="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.5/css/bootstrap.min.css" />
        <script type="application/javascript">window.appConfig = %s;</script>
        <script type="application/javascript" src="%s"></script>
    </head>
    <body><div id='app'></div></body>
</html>`

// AppConfig is the frontend configuration
type AppConfig struct {
	URI        string                      `json:"uri"`
	User       *session.User               `json:"user"`
	DefaultEnv string                      `json:"defaultEnv"`
	Envs       []*environments.Environment `json:"envs"`
}

func homeHandler(c echo.Context) error {
	if c.Get("user") == nil {
		staticLiveReload := config.GetBool(config.StaticLiveReload)
		return c.HTML(
			http.StatusOK,
			fmt.Sprintf(homeTemplate, "null", fmt.Sprintf("/static/app-%s.js", public.GetHash(staticLiveReload))),
		)
	}
	return appHandler(c)
}

func appHandler(c echo.Context) error {
	envs := environments.Environments()
	user, _ := c.Get("user").(*session.User)

	appConfig := AppConfig{
		URI:        config.GetString(config.URI),
		User:       user,
		DefaultEnv: config.GetString(config.DefaultEnv),
		Envs:       envs,
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
