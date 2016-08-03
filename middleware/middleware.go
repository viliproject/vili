package middleware

import (
	"gopkg.in/labstack/echo.v1"
)

// Func wraps the echo middleware func
type Func echo.MiddlewareFunc
