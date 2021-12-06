package controlplane

import (
	"fmt"
	"log"
	"time"

	"github.com/jasonlvhit/gocron"
	"github.com/patrickmn/go-cache"
	api "github.com/skillz/opvic/controlplane/api/v1alpha1"
	"github.com/skillz/opvic/utils"
)

const (
	// holds the list of all the agents that have been registered with the control plane
	AgentListCacheKey = "agents/list"
)

// Cach key for storing all the versions recieved from an agent in format:
// agents/:agentID
func AgentCacheKey(agentID string) string {
	return fmt.Sprintf("agents/%s", agentID)
}

// Cach key for each SubjectVersion recieved from an agent in format:
// :agentID/:versionID
func SubjectVersionCacheKey(agentID, versionID string) string {
	return fmt.Sprintf("%s/%s", agentID, versionID)
}

// Cach key for versions information of a SubjectVersion after getting the remote versions
// :agentID/:versionID/versions
func SubjectVersionInfoCacheKey(agentID, versionID string) string {
	return fmt.Sprintf("%s/%s/versions", agentID, versionID)
}

// Cach key that holds list IDs of all the subjects that an agent has sent to the control plane
// :agentID/list
func AgentSubjectVersionListCacheKey(agentID string) string {
	return fmt.Sprintf("%s/versions/list", agentID)
}

// SetAgentCache put the subjectVersions in the cache in the AgentCacheKey path
func (cp *ControlPlane) SetAgentCache(agentID string, subjectVersions api.SubjectVersions) {
	cp.cache.Set(AgentCacheKey(agentID), subjectVersions, cache.DefaultExpiration)
}

// GetAgentCache gets the SubjectVersions for an agent from the AgentCacheKey path cache
func (cp *ControlPlane) GetAgentCache(agentID string) (api.SubjectVersions, bool) {

	subjectVersions, found := cp.cache.Get(AgentCacheKey(agentID))
	if !found {
		return api.SubjectVersions{}, false
	}
	return subjectVersions.(api.SubjectVersions), true
}

// SetAgentCache sets the version in the agent payload in the cache
func (cp *ControlPlane) SetSubjectVersionCache(agentID, versionID string, subjectVersion api.SubjectVersion) {
	cp.cache.Set(SubjectVersionCacheKey(agentID, versionID), subjectVersion, cache.DefaultExpiration)
}

func (cp *ControlPlane) GetSubjectVersionCache(agent, versionID string) (api.SubjectVersion, bool) {
	subjectVersion, found := cp.cache.Get(SubjectVersionCacheKey(agent, versionID))
	if !found {
		return api.SubjectVersion{}, false
	}
	return subjectVersion.(api.SubjectVersion), true
}

func (cp *ControlPlane) SetSubjectVersionInfoCache(agentID, versionID string, versionInfo api.VersionInfos) {
	cp.cache.Set(SubjectVersionInfoCacheKey(agentID, versionID), versionInfo, cache.DefaultExpiration)
}

func (cp *ControlPlane) GetSubjectVersionInfoCache(agentID, versionID string) (api.VersionInfos, bool) {
	versionInfo, found := cp.cache.Get(SubjectVersionInfoCacheKey(agentID, versionID))
	if !found {
		return api.VersionInfos{}, false
	}
	return versionInfo.(api.VersionInfos), true
}

func (cp *ControlPlane) SetAgentListCache(agents api.Agents) {
	cp.cache.Set(AgentListCacheKey, agents, cache.DefaultExpiration)
}

func (cp *ControlPlane) GetAgentListCache() api.Agents {
	agents, found := cp.cache.Get(AgentListCacheKey)
	if !found {
		return api.Agents{}
	}
	return agents.(api.Agents)
}

// Check cache and update if necessary
func (cp *ControlPlane) UpdateAgentListCache(agentId string, agentTags map[string]string) {
	agents := cp.GetAgentListCache()
	found := false
	for _, agent := range agents {
		if agent.ID == agentId {
			found = true
			agent.Tags = agentTags
			agent.LastHeartbeat = time.Now().Unix()
		}
	}
	if !found {
		agents = append(agents, &api.Agent{
			ID:            agentId,
			Tags:          agentTags,
			LastHeartbeat: time.Now().Unix(),
		})
	}
	cp.SetAgentListCache(agents)
}

func (cp *ControlPlane) SetAgentSubjectVersionListCache(agentID string, list []string) {
	cp.cache.Set(AgentSubjectVersionListCacheKey(agentID), list, cache.DefaultExpiration)
}

func (cp *ControlPlane) GetAgentSubjectVersionListCache(agentID string) []string {
	list, found := cp.cache.Get(AgentSubjectVersionListCacheKey(agentID))
	if !found {
		return []string{}
	}
	return list.([]string)
}

func (cp *ControlPlane) UpdateAgentSubjectVersionsList(agentId, versionId string) {
	subjectList := cp.GetAgentSubjectVersionListCache(agentId)
	if !utils.Contains(subjectList, versionId) {
		subjectList = append(subjectList, versionId)
		cp.SetAgentSubjectVersionListCache(agentId, subjectList)
	}
}

func (cp *ControlPlane) CacheReconcile() {
	fmt.Println("Reconciling cache")
	cp.cache.DeleteExpired()
	cp.AgentListCacheReconcile()
	cp.AgentCacheReconcile()
	cp.SubjectVersionInfoCacheReconcile()
	fmt.Println("Cache reconciled")
}

func (cp *ControlPlane) AgentListCacheReconcile() {
	agents := cp.GetAgentListCache()
	newAgents := api.Agents{}
	for _, agent := range agents {
		if time.Now().Unix()-agent.LastHeartbeat < 3600 {
			newAgents = append(newAgents, agent)
		}
	}
	cp.SetAgentListCache(newAgents)
}

func (cp *ControlPlane) AgentCacheReconcile() {
	agents := cp.GetAgentListCache()
	for _, agent := range agents.ListIDs() {
		subjectVersions := []*api.SubjectVersion{}
		versionList := []string{}
		for _, versionID := range cp.GetAgentSubjectVersionListCache(agent) {
			if version, found := cp.GetSubjectVersionCache(agent, versionID); found {
				subjectVersions = append(subjectVersions, &version)
				versionList = append(versionList, versionID)
			}
		}
		cp.SetAgentCache(agent, subjectVersions)
		cp.SetAgentSubjectVersionListCache(agent, versionList)
	}
}

func (cp *ControlPlane) SubjectVersionInfoCacheReconcile() {
	agents := cp.GetAgentListCache()
	for _, agent := range agents.ListIDs() {
		if appvers, found := cp.GetAgentCache(agent); found {
			for _, ver := range appvers {
				verInfos, err := cp.GetSubjectVersionInfos(agent, ver)
				if err != nil {
					log.Printf("Error getting subject version info: %s", err)
					continue
				}
				cp.SetSubjectVersionInfoCache(agent, ver.ID, verInfos)
			}
		}
	}
}

func (cp *ControlPlane) executeCronJobs() {
	interval := uint64(cp.cacheReconcilerInterval.Seconds())
	gocron.Every(interval).Second().Do(cp.CacheReconcile)
	<-gocron.Start()
}
