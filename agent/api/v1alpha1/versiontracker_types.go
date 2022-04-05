/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type LocalStrategy string
type RemoteStrategy string

const (
	FieldSelection LocalStrategy = "FieldSelection"
	ImageTag       LocalStrategy = "ImageTag"

	HelmStrategyChartVersion RemoteStrategy = "chartVersion"
	HelmStrategyAppVersion   RemoteStrategy = "appVersion"
	GithubStrategyReleases   RemoteStrategy = "releases"
	GithubStrategyTags       RemoteStrategy = "tags"
)

var (
	ImageTagDefaults = LocalVersion{
		FieldSelector: ".spec.containers[0].image",
		Extraction: Extraction{
			Regex: Regex{
				Pattern: `.*:(.*)$`,
				Result:  "$1",
			},
		},
	}
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// VersionTrackerSpec defines the desired state of VersionTracker
type VersionTrackerSpec struct {
	// Important: Run "make" to regenerate code after modifying this file

	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:Required
	// Name of the app that is being tracked
	Name string `json:"name"`

	// +kubebuilder:validation:Required
	Resources Resources `json:"resources"`

	// +kubebuilder:validation:Required
	LocalVersion LocalVersion `json:"localVersion"`

	// +optional
	RemoteVersion RemoteVersion `json:"remoteVersion"`
}

type Resources struct {

	// +kubebuilder:default=Pods
	// +kubebuilder:validation:Enum = [Nodes, Pods, Deployments, DaemonSets, StatefulSets, ReplicaSets, CronJobs, Jobs]
	// Specifies the strategy to find the resources to track.(Default: `Pods`)
	// +optional
	Strategy string `json:"strategy"`

	// List of Namespaces to use when querying for resources (Default to query all namespaces)
	// +optional
	Namespaces []string `json:"namespaces"`

	// Label selector to use when querying for resources
	Selector *metav1.LabelSelector `json:"selector"`
}

type LocalVersion struct {
	// +kubebuilder:validation:Enum = ["ImageTag", "FieldSelection"]
	// +kubebuilder:default=ImageTag
	// +kubebuilder:validation:Required
	Strategy LocalStrategy `json:"strategy"`

	// Jsonpath to extract the version from the resource
	// +kubebuilder:Pattern=^.+$
	// +optional
	FieldSelector string `json:"fieldSelector"`

	// +optional
	Extraction Extraction `json:"extraction"`
}

type RemoteVersion struct {
	// +kubebuilder:validation:Enum = ["github", "helm-repo"]
	// +kubebuilder:default=github
	// +kubebuilder:validation:Required
	Provider string `json:"provider"`

	// +kubebuilder:validation:Enum = ["releases", "tags", "chartVersion", "appVersion"]
	// +kubebuilder:validation:Required
	Strategy RemoteStrategy `json:"strategy"`

	// Repository to get the remote version from.
	// e.g owner/repo or https://charts.bitnami.com/bitnami
	// +kubebuilder:validation:Required
	Repo string `json:"repo"`

	// Helm chart name to track. Required if `provider` is `helm-repo`
	// +optional
	Chart string `json:"chart,omitempty"`

	// +optional
	Extraction Extraction `json:"extraction"`

	// +optional
	Constraint string `json:"constraint,omitempty"`
}

type Extraction struct {
	// Regex to extract the version from the field
	// +optional
	Regex Regex `json:"regex,omitempty"`
}

type Regex struct {
	// Regex pattern to extract the version from the field
	// +kubebuilder:validation:Required
	Pattern string `json:"pattern"`

	// +kubebuilder:validation:Required
	// +kubebuilder:default=$1
	Result string `json:"result"`
}

type Version struct {
	ResourceCount int    `json:"resourceCount"`
	ResourceKind  string `json:"resourceKind"`
	ExtractedFrom string `json:"extractedFrom"`
	Version       string `json:"version"`
}

// VersionTrackerStatus defines the observed state of VersionTracker
type VersionTrackerStatus struct {
	ID                 *string        `json:"id,omitempty"`
	Namespace          *string        `json:"namespace,omitempty"`
	TotalResourceCount *int           `json:"totalResourceCount,omitempty"`
	UniqVersions       []*string      `json:"uniqVersions,omitempty"`
	Versions           []*Version     `json:"versions,omitempty"`
	LocalVersion       *LocalVersion  `json:"localVersion,omitempty"`
	RemoteVersion      *RemoteVersion `json:"remoteVersion,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// VersionTracker is the Schema for the versiontrackers API
type VersionTracker struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VersionTrackerSpec   `json:"spec,omitempty"`
	Status VersionTrackerStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// VersionTrackerList contains a list of VersionTracker
type VersionTrackerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VersionTracker `json:"items"`
}

func init() {
	SchemeBuilder.Register(&VersionTracker{}, &VersionTrackerList{})
}

func (v *VersionTracker) GetName() string {
	return v.Spec.Name
}

// GetKind returns the kind of the resource based on local version strategy
func (v *VersionTracker) GetResourceKind() string {
	return v.Spec.Resources.Strategy
}

func (v *VersionTracker) Validate() error {
	if v.Spec.LocalVersion.Strategy != ImageTag {
		if v.Spec.LocalVersion.FieldSelector == "" {
			return fmt.Errorf("fieldSelector is required when strategy is not ImageTag")
		}
	}
	return nil
}
func (v *VersionTracker) SetDefaults() VersionTracker {
	lv := v.Spec.LocalVersion
	if lv.Strategy == ImageTag {
		if lv.FieldSelector == "" {
			lv.FieldSelector = ImageTagDefaults.FieldSelector
		}
		if lv.Extraction.Regex.Pattern == "" {
			lv.Extraction.Regex.Pattern = ImageTagDefaults.Extraction.Regex.Pattern
		}
		if lv.Extraction.Regex.Result == "" {
			lv.Extraction.Regex.Result = ImageTagDefaults.Extraction.Regex.Result
		}
	}
	v.Spec.LocalVersion = lv
	return *v
}

func (v *VersionTracker) GetLocalVersion() LocalVersion {
	return v.Spec.LocalVersion
}

// GetObjectList returns client.ObjectList based on resource strategy
// It will be used to query for resources to track
func (v *VersionTracker) GetObjectList() (client.ObjectList, error) {
	switch v.Spec.Resources.Strategy {
	case "Nodes":
		return &corev1.NodeList{}, nil
	case "Pods":
		return &corev1.PodList{}, nil
	case "Deployments":
		return &appsv1.DeploymentList{}, nil
	case "DaemonSets":
		return &appsv1.DaemonSetList{}, nil
	case "StatefulSets":
		return &appsv1.StatefulSetList{}, nil
	case "ReplicaSets":
		return &appsv1.ReplicaSetList{}, nil
	case "CronJobs":
		return &batchv1.CronJobList{}, nil
	case "Jobs":
		return &batchv1.JobList{}, nil
	default:
		return nil, fmt.Errorf("unsupported resource type: %s", v.Spec.Resources.Strategy)
	}
}
