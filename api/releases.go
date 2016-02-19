package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/labstack/echo"

	"github.com/airware/vili/firebase"
	"github.com/airware/vili/session"
	"github.com/airware/vili/slack"
)

// Release is an object representing an image tag for release
type Release struct {
	URL      string    `json:"url,omitempty"`
	Time     time.Time `json:"time"`
	Username string    `json:"username"`
	Approved bool      `json:"approved"`
}

func releaseCreateHandler(c *echo.Context) error {
	app := c.Param("app")
	tag := c.Param("tag")
	release := &Release{}
	decoder := json.NewDecoder(c.Request().Body)
	err := decoder.Decode(release)
	if err != nil {
		return err
	}
	if release.URL != "" && !govalidator.IsURL(release.URL) {
		release.URL = ""
	}
	release.Time = time.Now()
	release.Username = c.Get("user").(*session.User).Username
	release.Approved = true
	err = firebase.Database().Child("releases").Child(app).Child(tag).Set(release)
	if err != nil {
		return err
	}
	slackMessage := fmt.Sprintf("*%s* tag *%s* approved for release by *%s*", app, tag, release.Username)
	if release.URL != "" {
		slackMessage += fmt.Sprintf(" - <%s|release notes>", release.URL)
	}
	err = slack.PostLogMessage(slackMessage, "info")
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, release)
}

func releaseDeleteHandler(c *echo.Context) error {
	app := c.Param("app")
	tag := c.Param("tag")
	err := firebase.Database().Child("releases").Child(app).Child(tag).Remove()
	if err != nil {
		return err
	}
	return c.String(http.StatusNoContent, "")
}
