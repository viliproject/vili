package middleware

import (
	"net"
	"time"

	"github.com/airware/vili/log"
	"github.com/labstack/echo"
)

// Logger logs relevant information about the request and the response
func Logger(name string) echo.MiddlewareFunc {
	return func(h echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			req := c.Request()
			res := c.Response()

			remoteAddr := req.RemoteAddr
			if ip := req.Header.Get(echo.XRealIP); ip != "" {
				remoteAddr = ip
			} else if ip = req.Header.Get(echo.XForwardedFor); ip != "" {
				remoteAddr = ip
			}
			remoteAddr, _, _ = net.SplitHostPort(remoteAddr)

			start := time.Now()
			if err := h(c); err != nil {
				c.Error(err)
			}
			stop := time.Now()
			path := req.URL.Path
			if path == "" {
				path = "/"
			}

			fields := log.Fields{
				"method":     req.Method,
				"proto":      req.Proto,
				"uri":        req.RequestURI,
				"path":       path,
				"remoteAddr": remoteAddr,
				"status":     res.Status(),
				"duration":   stop.Sub(start),
				"size":       res.Size(),
			}

			n := res.Status()
			switch {
			case n >= 500:
				log.WithFields(fields).Error(name + ".request")
			case n >= 400:
				log.WithFields(fields).Warn(name + ".request")
			case n == 204 && fields["path"] == "/admin/health":
				log.WithFields(fields).Debug(name + ".request")
			default:
				log.WithFields(fields).Info(name + ".request")
			}

			return nil
		}
	}
}
