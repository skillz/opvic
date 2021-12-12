package controlplane

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	api "github.com/skillz/opvic/controlplane/api/v1alpha1"
)

func (cp *ControlPlane) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		reqToken := c.Request.Header.Get("Authorization")
		if reqToken == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "missing authorization header",
			})
			return
		}
		token := strings.SplitN(reqToken, " ", 2)
		if len(token) != 2 || token[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid authorization header",
			})
			return
		}
		if token[1] != *cp.token {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid authorization token",
			})
			return
		}
		c.Next()
	}
}

func HeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Content-Type", "application/json")
		c.Writer.Header().Set("X-API-Version", api.APIVersion)
		c.Next()
	}
}

func (cp *ControlPlane) MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.String() == api.MetricsPath {
			c.Next()
			return
		} else if strings.HasPrefix(c.Request.URL.String(), "/assets") || strings.HasPrefix(c.Request.URL.String(), "/favicon.ico") {
			c.Next()
			return
		}

		c.Next()
		method := c.Request.Method
		path := c.Request.URL.String()
		status := strconv.Itoa(c.Writer.Status())

		cp.reqCount.WithLabelValues(method, path, status).Inc()
	}
}

func (cp *ControlPlane) RecoveryWithLogger(stack bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			logger := cp.logger
			if err := recover(); err != nil {
				// Check for a broken connection, as it is not really a
				// condition that warrants a panic stack trace.
				var brokenPipe bool
				if ne, ok := err.(*net.OpError); ok {
					if se, ok := ne.Err.(*os.SyscallError); ok {
						if strings.Contains(strings.ToLower(se.Error()), "broken pipe") || strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
							brokenPipe = true
						}
					}
				}

				httpRequest, _ := httputil.DumpRequest(c.Request, false)
				if brokenPipe {
					logger.Error(fmt.Errorf("%v", err), "broken pipe", "path", c.Request.URL.Path, "request", string(httpRequest))
					// If the connection is dead, we can't write a status to it.
					c.Error(err.(error)) // nolint: errcheck
					c.Abort()
					return
				}

				if stack {
					logger.Error(fmt.Errorf("%v", err), "[recovery fro panic]", "path", c.Request.URL.Path, "request", string(httpRequest), "stack", string(debug.Stack()))
				} else {
					logger.Error(fmt.Errorf("%v", err), "[recovery fro panic]", "path", c.Request.URL.Path, "request", string(httpRequest))
				}
				c.AbortWithStatus(http.StatusInternalServerError)
			}
		}()
		c.Next()
	}
}

func (cp *ControlPlane) LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := cp.logger
		start := time.Now()
		path := c.Request.URL.Path

		// Process request
		c.Next()

		if len(c.Errors) > 0 {
			// Append error field if this is an erroneous request.
			for _, e := range c.Errors.Errors() {
				logger.Error(fmt.Errorf("%v", e), "error", "path", path, "request", c.Request.URL.String())
			}
		} else {
			end := time.Now()
			logger.Info("request",
				"method", c.Request.Method,
				"path", path,
				"status", c.Writer.Status(),
				"latency", end.Sub(start),
				"ip", c.ClientIP(),
				"user-agent", c.Request.UserAgent(),
				"time", end,
			)
		}
	}
}
