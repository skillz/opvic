package controlplane

import (
	api "github.com/skillz/opvic/controlplane/api/v1alpha1"
	"github.com/skillz/opvic/controlplane/version"
	"github.com/skillz/opvic/utils"
)

func (cp *ControlPlane) GetSubjectVersionInfos(agentID string, ver *api.SubjectVersion) (api.VersionInfos, error) {
	log := cp.log.WithName("version").WithValues(
		"agent_id", agentID,
		"version_id", ver.ID,
		"provider", ver.RemoteVersion.Provider,
		"repo", ver.RemoteVersion.Repo,
	)
	log.V(1).Info("getting version infos")
	var latest string
	remoteversions, err := cp.provider.GetVersions(ver.RemoteVersion)
	if err != nil {
		log.Error(err, "Failed to get remote versions")
		return api.VersionInfos{}, err
	}
	subV, err := version.NewVersions("", remoteversions)
	if err != nil {
		return api.VersionInfos{}, err
	}
	if len(remoteversions) == 0 {
		log.V(1).Info("no remote version found. Is this expected? check the remoteVersion config")
		latest = "missing"
	} else {
		latest = subV.Latest().String()
	}

	verInfos := api.VersionInfos{
		ID:             ver.ID,
		AgentID:        agentID,
		ResourceCount:  ver.ResourceCount,
		LatestVersion:  latest,
		RemoteProvider: ver.RemoteVersion.Provider,
		RemoteRepo:     ver.RemoteVersion.Repo,
	}
	for _, v := range ver.Versions {
		subV.SetRunningVersion(v.RunningVersion)
		verInfos.Versions = append(verInfos.Versions, api.VersionInfo{
			RunningVersion:    subV.GetRunningVersion().String(),
			ResourceCount:     v.ResourceCount,
			ResourceKind:      v.ResourceKind,
			ExtractedFrom:     v.ExtractedFrom,
			LatestVersion:     latest,
			AvailableVersions: subV.GreaterThan().StringList(),
			AvailableMajors:   subV.MajorGreaterThan().StringList(),
			AvailableMinors:   subV.MinorGreaterThan().StringList(),
			AvailablePatches:  subV.PatchGreaterThan().StringList(),
			MajorAvailable:    subV.MajorAvailable(),
			MinorAvailable:    subV.MinorAvailable(),
			PatchAvailable:    subV.PatchAvailable(),
		})
		if !utils.Contains(verInfos.RunningVersions, v.RunningVersion) {
			verInfos.RunningVersions = append(verInfos.RunningVersions, v.RunningVersion)
		}
	}
	return verInfos, nil
}

func (cp *ControlPlane) GetAgentOverallVersionInfos(agentID string) ([]string, api.AgentVersionInfos) {
	subverIds := cp.GetAgentSubjectVersionListCache(agentID)
	versionIdList := []string{}
	var verInfos api.AgentVersionInfos
	for _, subverId := range subverIds {
		if verInfo, found := cp.GetSubjectVersionInfoCache(agentID, subverId); found {
			verInfos = append(verInfos, verInfo)
			versionIdList = append(versionIdList, verInfo.ID)
		}
	}
	return versionIdList, verInfos
}

func (cp *ControlPlane) GetOverallVersionInfos() []api.OverallVersionInfos {
	// get list of all agents
	agents := cp.GetAgentListCache()
	versionIDList := []string{}

	// Get version list and version infos for each agent
	for _, agentID := range agents.ListIDs() {
		versionIDList = append(versionIDList, cp.GetAgentSubjectVersionListCache(agentID)...)
	}
	// remove duplicates from version list
	versionIDList = utils.RemoveDuplicateStr(versionIDList)

	// Create overall version infos for each version from all agents
	overallVerInfos := make([]api.OverallVersionInfos, len(versionIDList))
	for i, versionID := range versionIDList {
		overallVerInfos[i] = api.OverallVersionInfos{
			versionID: []api.VersionInfos{},
		}
		for _, agentID := range agents.ListIDs() {
			if verInfo, found := cp.GetSubjectVersionInfoCache(agentID, versionID); found {
				overallVerInfos[i][versionID] = append(overallVerInfos[i][versionID], verInfo)
			}
		}
	}
	return overallVerInfos
}
