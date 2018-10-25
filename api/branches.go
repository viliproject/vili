package api

import (
	"net/http"

	"github.com/viliproject/vili/git"
	"github.com/labstack/echo"
)

func branchesGetHandler(c echo.Context) error {
	branches, err := git.Branches()
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string][]string{
		"branches": branches,
	})
}
