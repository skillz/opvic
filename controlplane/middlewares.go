package controlplane

import (
	"net/http"
	"strconv"
	"strings"

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
