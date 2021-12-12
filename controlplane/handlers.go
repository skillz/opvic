package controlplane

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	api "github.com/skillz/opvic/controlplane/api/v1alpha1"
)

// Metrics handler
func PrometheusHandler() gin.HandlerFunc {
	h := promhttp.Handler()

	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

// API handlers

// AgentsPost handles POST requests to /agents
func (cp *ControlPlane) AgentsPost() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ap api.AgentPayload
		if err := c.ShouldBindJSON(&ap); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusAccepted, gin.H{"message": "data received"})
		cp.log.Info(
			"received agent payload",
			"agent_id", ap.AgentID,
			"version_id", ap.Version.ID,
		)
		go func() {
			cp.UpdateAgentListCache(ap.AgentID, ap.AgentTags)
			cp.UpdateAgentSubjectVersionsList(ap.AgentID, ap.Version.ID)
			cp.SetSubjectVersionCache(ap.AgentID, ap.Version.ID, ap.Version)
		}()
	}
}

// AgentsGet handles GET requests to /agents
func (cp *ControlPlane) AgentsGet() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, cp.GetAgentListCache())
	}
}

// AgentGet handles GET requests to /agents/:id
func (cp *ControlPlane) AgentGet() gin.HandlerFunc {
	return func(c *gin.Context) {
		agentID := c.Param("id")
		if subjectVersion, found := cp.GetAgentCache(agentID); found {
			c.JSON(http.StatusOK, subjectVersion)
		} else {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		}
	}
}

// AgentsSubjectVersionGet handles GET requests to /agents/:id/versionId:
func (cp *ControlPlane) AgentsSubjectVersionGet() gin.HandlerFunc {
	return func(c *gin.Context) {
		agentID := c.Param("id")
		versionID := c.Param("versionId")
		if subjectVersion, found := cp.GetSubjectVersionCache(agentID, versionID); found {
			c.JSON(http.StatusOK, subjectVersion)
		} else {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		}
	}
}

// AgentsSubjectVersionGet handles GET requests to /agents/:id/versionId:/versions
func (cp *ControlPlane) AgentsSubjectVersionsInfoGet() gin.HandlerFunc {
	return func(c *gin.Context) {
		agentID := c.Param("id")
		versionID := c.Param("versionId")

		subjectVersionInfos, found := cp.GetSubjectVersionInfoCache(agentID, versionID)
		if !found {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		} else {
			c.JSON(http.StatusOK, subjectVersionInfos)
		}
	}
}

// OverviewGet handles GET requests to /overview
func (cp *ControlPlane) OverviewGet() gin.HandlerFunc {
	return func(c *gin.Context) {
		oveview := cp.GetOverallVersionInfos()
		c.JSON(http.StatusOK, oveview)
	}
}
