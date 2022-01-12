# Opvic
A tool to track and compare the versions of resources running in Kubernetes clusters and their latest available versions from remote sources (e.g. Github releases, Helm repositories, etc.)

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Table of Contents**

- [Opvic](#opvic)
  - [Motivation](#motivation)
  - [Architecture Design](#architecture-design)
    - [Agent](#agent)
      - [App Discovery](#app-discovery)
    - [Control Plane](#control-plane)
  - [Installation](#installation)
  - [Examples](#examples)
    - [Example 1 : Tracking CoreDNS From Container Image Tag](#example-1--tracking-coredns-from-container-image-tag)
    - [Example 2: Extract the Version From Any Field](#example-2-extract-the-version-from-any-field)
    - [Example 3: Use appVersion of a Helm Repository](#example-3-use-appversion-of-a-helm-repository)
    - [Example 4: Track your Helm Chart Versions](#example-4-track-your-helm-chart-versions)
  - [Development](#development)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## Motivation

Kubernetes is complicated. It becomes more so when there are hundreds, if not thousands, of addons and applications deployed on it. The problem is, how can we stay on top of new releases for all the addons that are deployed on Kubernetes? It’s not an out-of-box feature and we didn’t find any existing solutions for this purpose. GitHub lets us subscribe to repositories, thenemails us when there are new releases. But there are some flaws with this strategy:
- Not all projects are using Github Release and not all projects are on Github
- The first question we often ask is: great there’s now a new version, but which version are we currently running and what’s the difference between them?
- Email notifications can quickly go out of control if there are many addons watched
- Overheads for automation solutions like upgrading automatically if certain conditions are met.

We need a cloud-native solution that can track new versions from different release channels, detect the currently running versions, aggregate and present the version gaps, and enable further automation solutions.

## Architecture Design

Opvic project started with scalability in mind and that’s why we chose to decouple the process of tracking the running versions, configuration and getting the remote versions.

![Architecture Diagram](/images/diagram.png)

### Agent

Agents are installed on all clusters and send the collected information to the centralized control plane

- Agents interact with K8s API to retrieve figure out the running version of each component
- App discovery and configuration can be managed by custom resources (CRD)
- Each agent has a unique identifier (normally your cluster identifier)

#### App Discovery

By default, agents do not perform any action as it acts as a Kubernetes operator that watches certain custom resources. CRDs are highly flexible in terms of configuration as we support more components to track

Based on the CRD, agents can discover Kubernetes resources such as Nodes, Deployments, Pods, etc. based on Kubernetes standard [label selector](https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#resources-that-support-set-based-requirements) configuration. Then agents can extract the [semver](https://semver.org/) formatted versions from any field like image, labels or annotations.

### Control Plane
For the initial implementation the control plane will receive information from agents and store them in memory cache. Then it aggregates the information and retrieves the remote versions based on the configuration sent by each agent.
- Exposes an API endpoint for agents to send the collected information.
- Store the information in memory cache with a configurable expiration duration
- Interact with external systems such as Github, Helm registries, etc. to retrieve the versions between the running version and latest.
- Exposes Prometheus format metrics to show running versions across all clusters as well as available major, minor and patches versions to upgrade
- The API also exposes endpoints to query detailed information about each component


## Installation

There is a Helm chart available for Opvic [here](https://github.com/skillz/opvic/tree/main/charts/opvic). Begin by adding the repository:

```shell
helm repo add opvic https://skillz.github.io/opvic
```

Then deploy it with minimal required configuration. Check out [values.yaml](https://github.com/skillz/opvic/blob/main/charts/opvic/values.yaml) for all the available options:

```shell
helm upgrade --install --namespace opvic --create-namespace opvic opvic/opvic \
--set controlplane.enabled=true \
--set agent.enabled=true --set agent.identifier=test-agent \
--set sharedAuthentication.token=test
```

Ensure that pods are running and that you connect to the control plane:

```
kubectl get pods -n opvic

NAME                                   READY   STATUS    RESTARTS   AGE
opvic-agent-9b64fb88c-qfvnv            1/1     Running   0          34m
opvic-control-plane-56d6955f7d-wgttd   1/1     Running   0          34m

kubectl port-forward deployment/opvic-control-plane 8080:8080 -n opvic

Forwarding from 127.0.0.1:8080 -> 8080
Forwarding from [::1]:8080 -> 8080

curl -H "Authorization: Bearer test" localhost:8080/api/v1alpha1/ping
pong
```

Notes:
- You only need to run the control plane in one of your clusters
- You need to deploy the agent and CRDs in all clusters
- You don’t need an ingress if the control plane and agent run on the same cluster.
- VersionTracker resources should be deployed in all clusters

## Examples

Below is an example of a VersionTracker resource:

```yaml
apiVersion: opvic.skillz.com/v1alpha1
kind: VersionTracker
metadata:
  name: myApp # name of the subject
spec:
  name: myApp # Unique identifier of the app to track
  resources: # How agent should find the resources to extract the version
    strategy: Pods # Resource Kind. Pods, Nodes, Deployments, etc (default to Pods)
    namespaces: # (optional) namespaces of the app (default to query all namespaces)
      - kube-system
    selector: # Kubernetes standard selector configuration
      matchLabels:
        app.kubernetes.io/name: myApp
  localVersion: # How agent should extract the version
   strategy: FieldSelection # strategy to use for the app (ImageTag, FieldSelection)
   fieldSelector: '.metadata.labels.myApp\.io/version' # Valid JsonPath to extract the version from the resource
   extraction: # Regex to extract the version from the resource
     regex:
       pattern: '^v([0-9]+\.[0-9]+\.[0-9]+)$'
       result: '$1'
  remoteVersion: # How control plane should find the remote versions
    provider: github # name of the provider (github, helm)
    strategy: releases # method to use to get the remote versions (releases, tags)
    repo: owner/repoName # name of the repository (owner/repoName)
    extraction:
     regex:
       pattern: '^myApp-v([0-9]+\.[0-9]+\.[0-9]+)$'
       result: '$1'
    constraint: '~>3' # Semver constraint to use to filter the remote versions
```

Note that if the remote versions are not exposed or the provider is not supported by Opvic yet, you can still track the running versions and not specify the remoteVersion configuration.

### Example 1 : Tracking CoreDNS From Container Image Tag

Let’s say you want to track the running version of CoreDNS on your cluster as well as the available upstream versions. You can deploy a VersionTracker resource `coredns.yaml` like this:

```yaml
apiVersion: opvic.skillz.com/v1alpha1
kind: VersionTracker
metadata:
  name: coredns
spec:
  name: coredns
  resources:
    namespaces:
      - kube-system
    selector:
      matchLabels:
        k8s-app: kube-dns
  localVersion:
   strategy: ImageTag
  remoteVersion:
   provider: github
   strategy: releases
   repo: coredns/coredns
   extraction:
     regex:
       pattern: ^v([0-9]+\.[0-9]+\.[0-9]+)$
       result: $1
```

Let's apply this:

```shell
kubectl apply  -f coredns.yaml -n opvic
```

Since most versions can be extracted from containers’ image tags, you can use the **ImageTag** strategy which extracts the version from the first container image tag of the resource.

For remote versions, you can use the **github** provider and look at releases by using **releases** strategy. You need to specify the github repository and a regex for extraction.

Now you can query the control plane for running versions:

```shell
curl -H "Authorization: Bearer test" localhost:8080/api/v1alpha1/agents/test/coredns | jq
```

And you would get a response like this:

```json
{
 "id": "coredns",
 "namespace": "opvic",
 "count": 1,
 "uniqVersions": [
   "1.7.0"
 ],
 "versions": [
   {
     "runningVersion": "1.7.0",
     "resourceCount": 1,
     "resourceKind": "Pods",
     "extractedFrom": "k8s.gcr.io/coredns:1.7.0"
   }
 ],
 "remoteVersion": {
   "provider": "github",
   "strategy": "releases",
   "repo": "coredns/coredns",
   "extraction": {
     "regex": {
       "pattern": "^v([0-9]+\\.[0-9]+\\.[0-9]+)$",
       "result": "$1"
     }
   }
 }
}
```

Normally it takes up to a minute for the control plane to aggregate collected data and retrieve remote versions since that process runs on an interval in the background. To query the versions:

```shell
curl -H "Authorization: Bearer test" localhost:8080/api/v1alpha1/agents/test/coredns/versions | jq
```

And you would get a response like this:

```json
{
 "id": "coredns",
 "agentId": "test",
 "resourceCount": 1,
 "runningVersions": [
   "1.7.0"
 ],
 "latestVersion": "1.8.6",
 "remoteProvider": "github",
 "remoteRepo": "coredns/coredns",
 "versions": [
   {
     "currentVersion": "1.7.0",
     "resourceCount": 1,
     "resourceKind": "Pods",
     "extractedFrom": "k8s.gcr.io/coredns:1.7.0",
     "latestVersion": "1.8.6",
     "availableVersions": [
       "1.7.1",
       "1.8.0",
       "1.8.1",
       "1.8.2",
       "1.8.3",
       "1.8.4",
       "1.8.5",
       "1.8.6"
     ],
     "availableMajors": [],
     "availableMinors": [
       "1.8.0",
       "1.8.1",
       "1.8.2",
       "1.8.3",
       "1.8.4",
       "1.8.5",
       "1.8.6"
     ],
     "availablePatches": [
       "1.7.1"
     ],
     "majorAvailable": false,
     "minorAvailable": true,
     "patchAvailable": true
   }
 ]
}
```

The control plane also exposes Prometheus metrics at `/metrics` endpoint, so let’s take a look at those:

```shell
# HELP opvic_controlplane_agent_last_heartbeat Last time the agent was seen
# TYPE opvic_controlplane_agent_last_heartbeat gauge
opvic_controlplane_agent_last_heartbeat{agent_id="test",tags=""} 1.639773192e+09
# HELP opvic_controlplane_major_versions_count Number of available major versions to upgrade to
# TYPE opvic_controlplane_major_versions_count gauge
opvic_controlplane_major_versions_count{agent_id="test",available_major_versions="",remote_provider="github",remote_repo="coredns/coredns",resource_kind="Pods",running_version="1.7.0",version_id="coredns"} 0
# HELP opvic_controlplane_minor_versions_count Number of available minor versions to upgrade to
# TYPE opvic_controlplane_minor_versions_count gauge
opvic_controlplane_minor_versions_count{agent_id="test",available_minor_versions="1.8.0,1.8.1,1.8.2,1.8.3,1.8.4,1.8.5,1.8.6",remote_provider="github",remote_repo="coredns/coredns",resource_kind="Pods",running_version="1.7.0",version_id="coredns"} 7
# HELP opvic_controlplane_patch_versions_count Number of available patch versions to upgrade to
# TYPE opvic_controlplane_patch_versions_count gauge
opvic_controlplane_patch_versions_count{agent_id="test",available_patch_versions="1.7.1",remote_provider="github",remote_repo="coredns/coredns",resource_kind="Pods",running_version="1.7.0",version_id="coredns"} 1
# HELP opvic_controlplane_requests_total The number of HTTP requests processed
# TYPE opvic_controlplane_requests_total counter
opvic_controlplane_requests_total{method="GET",path="/api/v1alpha1/agents/test/coredns",status="200"} 1
opvic_controlplane_requests_total{method="GET",path="/api/v1alpha1/agents/test/coredns/versions",status="200"} 1
opvic_controlplane_requests_total{method="GET",path="/api/v1alpha1/ping",status="200"} 1
opvic_controlplane_requests_total{method="POST",path="/api/v1alpha1/agents",status="202"} 15
# HELP opvic_controlplane_version_resource_count Number of resources running with a specific version
# TYPE opvic_controlplane_version_resource_count gauge
opvic_controlplane_version_resource_count{agent_id="test",extracted_from="k8s.gcr.io/coredns:1.7.0",latest_version="1.8.6",remote_provider="github",remote_repo="coredns/coredns",resource_kind="Pods",running_version="1.7.0",version_id="coredns"} 1
# HELP opvic_provider_github_rate_limit_remaining The number of requests remaining in the current rate limit window.
# TYPE opvic_provider_github_rate_limit_remaining gauge
opvic_provider_github_rate_limit_remaining 58
```

### Example 2: Extract the Version From Any Field

you can extract the local version from any field. So let’s grab Kubelet version of your nodes:

```yaml
kind: VersionTracker
metadata:
  name: kubernetes
spec:
  name: kubernetes
  resources:
    strategy: Nodes
    selector:
      matchExpressions:
      - key: kubernetes.io/hostname
        operator: Exists
  localVersion:
    strategy: FieldSelector
    fieldSelector: '.status.nodeInfo.kubeletVersion'
    extraction:
      regex:
        pattern: '^v([0-9]+\.[0-9]+\.[0-9]+)'
        result: $1
  remoteVersion:
    provider: github
    strategy: tags
    repo: kubernetes/kubernetes
    extraction:
      regex:
        pattern: ^v([0-9]+\.[0-9]+\.[0-9]+)
        result: $1
```

In this example we use **FieldSelector** strategy in the **localVersion** configuration and set the JsonPath to `.status.nodeInfo.kubeletVersion` to extract the Kubelet version from the node status. We also use **tags** strategy in the **remoteVersion** configuration to get the latest version from the remote repository.

### Example 3: Use appVersion of a Helm Repository

If the application that you are tracking does not have a GitHub repository, check to see if it has a Helm chart as appVersion in the helm charts normally represents the actual version. You can use **helm** provider and **appVersion** strategy:

```yaml
apiVersion: opvic.skillz.com/v1alpha1
kind: VersionTracker
metadata:
  name: artifactory
spec:
  name: artifactory
  resources:
    namespaces:
      - artifactory
    selector:
      matchLabels:
        component: artifactory
  localVersion:
    strategy: ImageTag
  remoteVersion:
    provider: helm
    strategy: appVersion
    repo: https://charts.jfrog.io
    chart: artifactory
```

### Example 4: Track your Helm Chart Versions

You can also track Helm chart versions using the **helm** provider and **chartVersion** strategy. Assuming the deployed helm chart version is exposed in one of the labels called chart:

```yaml
apiVersion: opvic.skillz.com/v1alpha1
kind: VersionTracker
metadata:
  name: artifactory-helm
spec:
  name: artifactory-helm-chart
  resources:
    namespaces:
      - artifactory
    selector:
      matchLabels:
        component: artifactory
  localVersion:
    strategy: FieldSelector
    fieldSelector: '.metadata.labels.chart'
    extraction:
      regex:
        pattern: ^artifactory-([0-9]+\.[0-9]+\.[0-9]+)$
        result: $1
  remoteVersion:
    provider: helm
    strategy: chartVersion
    repo: https://charts.jfrog.io
    chart: artifactory
```

## Development

Makefile is available in the repository. to see all the options available to you, run:

```shell
make help

Usage:
  make <target>
  help             Display this help.

Development
  manifests        Generate CustomResourceDefinition object and copy to the charts/opvic directory
  generate         Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
  fmt              Run go fmt against code.
  vet              Run go vet against code.
  test             Run tests.

Build
  build            Build control plane and agent binary.
  run              Run control plane from your host.
  run-agent        Run an agent from your host.
  docker-build     Build docker image for control plane and agent.

Deployment
  install          Install CRDs into the K8s cluster specified in ~/.kube/config.
  uninstall        Uninstall CRDs from the K8s cluster specified in ~/.kube/config.
  controller-gen   Download controller-gen locally if necessary.
  kustomize        Download kustomize locally if necessary.
```
