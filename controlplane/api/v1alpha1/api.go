package api

import (
	"fmt"
	"sort"

	"github.com/skillz/opvic/agent/api/v1alpha1"
)

const APIVersion = "v1alpha1"

const (
	MetricsPath = "/metrics"
	PingAPIPath = "/ping"

	// Agent endpoints
	AgentsAPIPath                = "/agents"
	AgentAPIPath                 = "/agents/:id"
	AgentsSubjectVersionPath     = "/agents/:id/:versionId"
	AgentsSubjectVersionInfoPath = "/agents/:id/:versionId/versions"
)

var (
	APIGroup                         = fmt.Sprintf("/api/%s", APIVersion)
	PingAPIEndpoint                  = GetAPIEndpoint(PingAPIPath)
	AgentsAPIEndpoint                = GetAPIEndpoint(AgentsAPIPath)
	AgentAPIEndpoint                 = GetAPIEndpoint(AgentAPIPath)
	AgentsSubjectVersionEndpoint     = GetAPIEndpoint(AgentsSubjectVersionPath)
	AgentsSubjectVersionInfoEndpoint = GetAPIEndpoint(AgentsSubjectVersionInfoPath)
)

// gets the end point in `/<path>` format and returns (/api/<version>/<endpoint>)
func GetAPIEndpoint(endpoint string) string {
	return fmt.Sprintf("%s%s", APIGroup, endpoint)
}

type Agent struct {
	ID            string            `json:"id"`
	Tags          map[string]string `json:"tags"`
	LastHeartbeat int64             `json:"lastHeartbeat"`
}

type Agents []*Agent

// returns the sorted list of agent IDs
func (a *Agents) ListIDs() []string {
	var list []string
	for _, agent := range *a {
		list = append(list, agent.ID)
	}
	sort.Strings(list)
	return list
}

// Payload for the /agents endpoint
type AgentPayload struct {
	// Identifier of the agent
	AgentID string `json:"agentId" binding:"required"`
	// Tags associated with the agent
	AgentTags map[string]string `json:"agentTags"`
	// Version information collected by the agent
	Version SubjectVersion `json:"version" binding:"required"`
}

// SubjectVersion contains all versions collected for a subject
type SubjectVersion struct {
	// Identifier of the subject
	ID string `json:"id" binding:"required"`
	// NameSpace of the CRD
	NameSpace string `json:"namespace" binding:"required"`
	// Total Number of resources collected
	ResourceCount int `json:"count" binding:"required"`
	// List of running versions
	RunningVersions []string `json:"uniqVersions" binding:"required"`
	// List of versions collected for the subject
	Versions []Version `json:"versions" binding:"required"`
	// Information for getting the remote version
	RemoteVersion v1alpha1.RemoteVersion `json:"remoteVersion"`
}

// SubjectVersions is a list of SubjectVersion
type SubjectVersions []*SubjectVersion

// Version represents a version of a subject
type Version struct {
	// Runtime version extracted by the agent
	RunningVersion string `json:"runningVersion"`
	// Number of resources running with the version
	ResourceCount int `json:"resourceCount"`
	// Resource Kind (e.g. "Nodes, "Pods", "Deployments")
	ResourceKind string `json:"resourceKind"`
	// Field value that version is extracted from
	ExtractedFrom string `json:"extractedFrom"`
}

// VersionInfo contains information the running and remote versions of a subject
type VersionInfo struct {
	// Running version of the subject that was reported
	RunningVersion string `json:"currentVersion"`
	// Number of resources running with the version
	ResourceCount int `json:"resourceCount"`
	// Resource Kind (e.g. "Nodes, "Pods", "Deployments")
	ResourceKind string `json:"resourceKind"`
	// Field value that the version is extracted from
	ExtractedFrom string `json:"extractedFrom"`
	// Latest version of the remote version
	LatestVersion string `json:"latestVersion"`
	// List of all available versions above the running version
	AvailableVersions []string `json:"availableVersions"`
	// List of all available major versions above the running version
	AvailableMajors []string `json:"availableMajors"`
	// List of all available minor versions above the running version
	AvailableMinors []string `json:"availableMinors"`
	// List of all available patch versions above the running version
	AvailablePatches []string `json:"availablePatches"`
	// Boolean indicating if a newer major version is available
	MajorAvailable bool `json:"majorAvailable"`
	// Boolean indicating if a newer minor version is available
	MinorAvailable bool `json:"minorAvailable"`
	// Boolean indicating if a newer patch version is available
	PatchAvailable bool `json:"patchAvailable"`
}

// VersionInfos holds all the information on a subject version
type VersionInfos struct {
	// Identifier of the subject
	ID string `json:"id"`
	// Agent that reported the version
	AgentID string `json:"agentId"`
	// Total number of resources collected
	ResourceCount int `json:"resourceCount"`
	// List of all running versions
	RunningVersions []string `json:"runningVersions"`
	// Latest version based on the remote provider configuration
	LatestVersion string `json:"latestVersion"`
	// Remote provider for extracting remote versions
	RemoteProvider string `json:"remoteProvider"`
	// Remote repository or extracting remote versions
	RemoteRepo string `json:"remoteRepo"`
	// List of all VersionInfos collected for the subject
	Versions []VersionInfo `json:"versions"`
}

type AgentVersionInfos []VersionInfos

func (a *AgentVersionInfos) VersionIDList() []string {
	var versionIDs []string
	for _, v := range *a {
		versionIDs = append(versionIDs, v.ID)
	}
	sort.Strings(versionIDs)
	return versionIDs
}

// OverallVersionInfos has unique version information from all agentss
type OverallVersionInfos map[string][]VersionInfos
