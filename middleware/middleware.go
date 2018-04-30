package middleware

import (
	"github.com/labstack/echo"
)

// Func wraps the echo middleware func
type Func echo.MiddlewareFunc
