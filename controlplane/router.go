package controlplane

import (
	"github.com/gin-gonic/gin"
	api "github.com/skillz/opvic/controlplane/api/v1alpha1"
)

func (cp *ControlPlane) SetupRouter() *gin.Engine {
	r := gin.Default()

	// Add metrics middleware
	r.Use(cp.MetricsMiddleware())

	// Remove the extra slash anywhere in the path (e.g. /api/v1//foo -> /api/v1/foo)
	r.RemoveExtraSlash = true

	// Add AuthMiddleware to all API routes
	v1alpha1 := r.Group(api.APIGroup).Use(cp.AuthMiddleware())

	// Metrics router
	r.GET(api.MetricsPath, PrometheusHandler())

	// Ping router
	v1alpha1.GET(api.PingAPIPath, func(c *gin.Context) { c.String(200, "pong") })

	// Agents router
	v1alpha1.POST(api.AgentsAPIPath, cp.AgentsPost())
	v1alpha1.GET(api.AgentsAPIPath, cp.AgentsGet())
	v1alpha1.GET(api.AgentAPIPath, cp.AgentGet())
	v1alpha1.GET(api.AgentsSubjectVersionPath, cp.AgentsSubjectVersionGet())
	v1alpha1.GET(api.AgentsSubjectVersionInfoPath, cp.AgentsSubjectVersionsInfoGet())

	return r
}
