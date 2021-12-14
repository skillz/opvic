package controlplane

import (
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	metricNamespace = "opvic"
	metricSubsystem = "controlplane"
)

var (
	commonLabels = []string{"version_id", "agent_id", "running_version", "resource_kind", "remote_provider", "remote_repo"}
)

func newMetric(metricName string, docString string, commonLabelsNames []string, labelNames []string) *prometheus.Desc {
	labels := append(commonLabelsNames, labelNames...)
	return prometheus.NewDesc(prometheus.BuildFQName(metricNamespace, metricSubsystem, metricName), docString, labels, nil)
}

var (
	versionResouceCountMetric   = newMetric("version_resource_count", "Number of resources running with a specific version", commonLabels, []string{"extracted_from", "latest_version"})
	availableMajorVersionMetric = newMetric("major_versions_count", "Number of available major versions to upgrade to", commonLabels, []string{"available_major_versions"})
	availableMinorVersionMetric = newMetric("minor_versions_count", "Number of available minor versions to upgrade to", commonLabels, []string{"available_minor_versions"})
	availablePatchVersionMetric = newMetric("patch_versions_count", "Number of available patch versions to upgrade to", commonLabels, []string{"available_patch_versions"})

	agentMetric = newMetric("agent_last_heartbeat", "Last time the agent was seen", []string{}, []string{"agent_id", "tags"})
)

func (cp *ControlPlane) Describe(ch chan<- *prometheus.Desc) {
	ch <- versionResouceCountMetric
	ch <- availableMajorVersionMetric
	ch <- availableMinorVersionMetric
	ch <- availablePatchVersionMetric
}

func (cp *ControlPlane) Collect(ch chan<- prometheus.Metric) {
	cp.mutex.Lock()
	defer cp.mutex.Unlock()

	cp.setMetrics(ch)
}

func (cp *ControlPlane) setMetrics(ch chan<- prometheus.Metric) {
	cp.setVersionMetrics(ch)
	cp.setAgentMetrics(ch)
}

func (cp *ControlPlane) setVersionMetrics(ch chan<- prometheus.Metric) {
	aggregatedData := cp.GetOverallVersionInfos()
	for _, overallVersionInfos := range aggregatedData {
		for _, overallVersionInfo := range overallVersionInfos {
			for _, versionInfos := range overallVersionInfo {
				for _, v := range versionInfos.Versions {
					// resource count of each version
					ch <- prometheus.MustNewConstMetric(
						versionResouceCountMetric,
						prometheus.GaugeValue,
						float64(v.ResourceCount),
						versionInfos.ID,
						versionInfos.AgentID,
						v.RunningVersion,
						v.ResourceKind,
						versionInfos.RemoteProvider,
						versionInfos.RemoteRepo,
						v.ExtractedFrom,
						v.LatestVersion,
					)

					if v.LatestVersion == MissingLatest {
						continue
					}
					// Metrics for available major,minor and patch versions
					ch <- prometheus.MustNewConstMetric(
						availableMajorVersionMetric,
						prometheus.GaugeValue,
						float64(len(v.AvailableMajors)),
						versionInfos.ID,
						versionInfos.AgentID,
						v.RunningVersion,
						v.ResourceKind,
						versionInfos.RemoteProvider,
						versionInfos.RemoteRepo,
						strings.Join(v.AvailableMajors, ","),
					)
					ch <- prometheus.MustNewConstMetric(
						availableMinorVersionMetric,
						prometheus.GaugeValue,
						float64(len(v.AvailableMinors)),
						versionInfos.ID,
						versionInfos.AgentID,
						v.RunningVersion,
						v.ResourceKind,
						versionInfos.RemoteProvider,
						versionInfos.RemoteRepo,
						strings.Join(v.AvailableMinors, ","),
					)
					ch <- prometheus.MustNewConstMetric(
						availablePatchVersionMetric,
						prometheus.GaugeValue,
						float64(len(v.AvailablePatches)),
						versionInfos.ID,
						versionInfos.AgentID,
						v.RunningVersion,
						v.ResourceKind,
						versionInfos.RemoteProvider,
						versionInfos.RemoteRepo,
						strings.Join(v.AvailablePatches, ","),
					)
				}
			}
		}
	}
}

func (cp *ControlPlane) setAgentMetrics(ch chan<- prometheus.Metric) {
	agents := cp.GetAgentListCache()
	for _, agent := range agents {
		var tags string
		for key, value := range agent.Tags {
			// don't add , if there is no other tag
			if tags != "" {
				tags += ","
			}
			tags += key + "=" + value
		}

		ch <- prometheus.MustNewConstMetric(
			agentMetric,
			prometheus.GaugeValue,
			float64(agent.LastHeartbeat),
			agent.ID,
			tags,
		)
	}
}
