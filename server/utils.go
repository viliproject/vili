package server

import (
	"github.com/airware/vili/errors"
	"github.com/labstack/echo"
)

// ErrorResponse returns a JSON error response
func ErrorResponse(c *echo.Context, e *errors.ErrorResponse) error {
	return c.JSON(e.Status, e)
}
