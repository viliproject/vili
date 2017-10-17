package api

import (
	"net/http"

	"github.com/airware/vili/git"
	echo "gopkg.in/labstack/echo.v1"
)

func branchesGetHandler(c *echo.Context) error {
	branches, err := git.Branches()
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string][]string{
		"branches": branches,
	})
}
