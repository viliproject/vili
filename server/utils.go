package server

import (
	"github.com/airware/vili/errors"
	"gopkg.in/labstack/echo.v1"
)

// ErrorResponse returns a JSON error response
func ErrorResponse(c *echo.Context, e *errors.ErrorResponse) error {
	return c.JSON(e.Status, e)
}
