package agent

import (
	"fmt"
	"strings"

	"github.com/skillz/opvic/agent/api/v1alpha1"
	"github.com/skillz/opvic/utils"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type SubjectVersion struct {
	ID                 string
	Namespace          string
	TotalResourceCount int
	UniqVersions       []string
	Versions           []*Version
	RemoteVersion      v1alpha1.RemoteVersion
}

type Version struct {
	ResourceCount int
	ResourceKind  string
	ExtractedFrom string
	Version       string
}

// ExtractSubjectVersion looks at the feild of each individuel resource and extracts the version
// based on the extraction configuration in the VersionTracker
func (r *VersionTrackerReconciler) ExtractSubjectVersion(v v1alpha1.VersionTracker, items []interface{}) SubjectVersion {
	log := r.Log.WithName("extractor").WithValues("VersionTracker", fmt.Sprintf("%s/%s", v.ObjectMeta.Namespace, v.ObjectMeta.Name))
	var version string
	var versions []string
	uniqueVersions := []string{}
	lv := v.GetLocalVersion()

	if len(items) == 0 {
		log.Info("no resource was found. skipping version extraction")
		return SubjectVersion{}
	}

	appVersion := &SubjectVersion{
		ID:            v.Spec.Name,
		Namespace:     v.ObjectMeta.Namespace,
		RemoteVersion: v.Spec.RemoteVersion,
	}

	log.V(1).Info("resource count", "count", len(items))
	for _, i := range items {
		valueStrings, err := getFeilds(lv.FieldSelector, i)
		if err != nil {
			log.Error(err, "failed to get fields from the resource")
			reconciliationErrorsTotal.Inc()
			continue
		}
		if len(valueStrings) == 0 || len(valueStrings) > 1 {
			log.Error(fmt.Errorf("jsonpath returned unexpected number of values: %d", len(valueStrings)), "unexpected number of values", "fieldSelector", lv.FieldSelector)
			reconciliationErrorsTotal.Inc()
			continue
		}
		fieldValue := valueStrings[0]
		version = GetResultsFromRegex(lv.Extraction.Regex.Pattern, lv.Extraction.Regex.Result, fieldValue)
		if version == "" {
			log.Error(fmt.Errorf("failed to extract version from: %s", fieldValue), "extraction failed", "regex", lv.Extraction.Regex.Pattern, "result template", lv.Extraction.Regex.Result)
			reconciliationErrorsTotal.Inc()
			continue
		}

		// add the version to the list of unique versions if it's not already there
		if !utils.Contains(uniqueVersions, version) {
			uniqueVersions = append(uniqueVersions, version)
			appVersion.Versions = append(appVersion.Versions, &Version{
				Version:       version,
				ExtractedFrom: fieldValue,
				ResourceKind:  v.GetResourceKind(),
			})
		}
		versions = append(versions, version)
	}
	appVersion.TotalResourceCount = len(items)
	appVersion.UniqVersions = uniqueVersions
	// Set the number of pods for each version
	for _, t := range appVersion.Versions {
		for _, v := range versions {
			if t.Version == v {
				t.ResourceCount++
			}
		}
	}
	if len(appVersion.Versions) == 0 {
		log.Info("could not extract any versions from the resources")
	} else {
		log.Info("unique version(s)", "version(s)", strings.Join(uniqueVersions, ", "))
		for _, v := range appVersion.Versions {
			log.V(1).Info("extracted version", "version", v.Version, "resource count", v.ResourceCount)
		}
	}
	return *appVersion
}

// Returns the list of items from the resources based on the resource type
func GetItems(resources client.ObjectList) []interface{} {
	var items []interface{}
	switch resources.(type) {
	case *corev1.NodeList:
		items = make([]interface{}, len(resources.(*corev1.NodeList).Items))
		for i, v := range resources.(*corev1.NodeList).Items {
			items[i] = v
		}
		return items
	case *corev1.PodList:
		items = make([]interface{}, len(resources.(*corev1.PodList).Items))
		for i, item := range resources.(*corev1.PodList).Items {
			items[i] = item
		}
		return items
	case *appsv1.DeploymentList:
		items = make([]interface{}, len(resources.(*appsv1.DeploymentList).Items))
		for i, item := range resources.(*appsv1.DeploymentList).Items {
			items[i] = item
		}
		return items
	case *appsv1.DaemonSetList:
		items = make([]interface{}, len(resources.(*appsv1.DaemonSetList).Items))
		for i, item := range resources.(*appsv1.DaemonSetList).Items {
			items[i] = item
		}
		return items
	case *appsv1.ReplicaSetList:
		items = make([]interface{}, len(resources.(*appsv1.ReplicaSetList).Items))
		for i, item := range resources.(*appsv1.ReplicaSetList).Items {
			items[i] = item
		}
		return items
	case *appsv1.StatefulSetList:
		items = make([]interface{}, len(resources.(*appsv1.StatefulSetList).Items))
		for i, item := range resources.(*appsv1.StatefulSetList).Items {
			items[i] = item
		}
		return items
	case *batchv1.CronJobList:
		items = make([]interface{}, len(resources.(*batchv1.CronJobList).Items))
		for i, item := range resources.(*batchv1.CronJobList).Items {
			items[i] = item
		}
		return items
	case *batchv1.JobList:
		items = make([]interface{}, len(resources.(*batchv1.JobList).Items))
		for i, item := range resources.(*batchv1.JobList).Items {
			items[i] = item
		}
		return items
	}
	return nil
}
