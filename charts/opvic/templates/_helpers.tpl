{{/*
Expand the name of the chart.
*/}}
{{- define "opvic.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "opvic.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "opvic.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "opvic.labels" -}}
helm.sh/chart: {{ include "opvic.chart" . }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Common Selector labels
*/}}
{{- define "opvic.selectorLabels" -}}
app.kubernetes.io/name: {{ include "opvic.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Shared Auth secret name
*/}}
{{- define "opvic.sharedAuthSecretName" -}}
{{- if or .Values.controlplane.enabled .Values.agent.enabled }}
{{- if .Values.sharedAuthentication.token }}
{{- printf "%s-shared-auth-token" (include "opvic.fullname" .) }}
{{- else if .Values.sharedAuthentication.existingSecret }}
{{- .Values.sharedAuthentication.existingSecret }}
{{- else }}
{{- fail "sharedAuthentication.token or sharedAuthentication.existingSecret must be set" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Control plane labels
*/}}
{{- define "opvic.controlplane.labels" -}}
{{ include "opvic.labels" . }}
{{ include "opvic.controlplane.selectorLabels" . }}
{{- end }}

{{/*
Control plane Selector labels
*/}}
{{- define "opvic.controlplane.selectorLabels" -}}
{{ include "opvic.selectorLabels" . }}
app.kubernetes.io/component: control-plane
{{- end }}

{{/*
Control plane service account to use
*/}}
{{- define "opvic.controlplane.serviceAccountName" -}}
{{- if .Values.controlplane.serviceAccount.create }}
{{- default (printf "%s-control-plane" (include "opvic.fullname" .)) .Values.controlplane.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.controlplane.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Control Plane Githu Provider secret name
*/}}
{{- define "opvic.controlplane.providers.github.secretName" -}}
{{- if .Values.controlplane.providers.github.createSecret }}
{{- printf "%s-control-plane-provider-github-secrets" (include "opvic.fullname" .)}}
{{- else if .Values.controlplane.providers.github.existingSecret }}
{{- .Values.sharedAuthentication.existingSecret }}
{{- end }}
{{- end }}

{{/*
Agent labels
*/}}
{{- define "opvic.agent.labels" -}}
{{ include "opvic.labels" . }}
{{ include "opvic.agent.selectorLabels" . }}
{{- end }}

{{/*
Agent Selector labels
*/}}
{{- define "opvic.agent.selectorLabels" -}}
{{ include "opvic.selectorLabels" . }}
app.kubernetes.io/component: agent
{{- end }}

{{/*
Agent service account to use
*/}}
{{- define "opvic.agent.serviceAccountName" -}}
{{- if .Values.agent.serviceAccount.create }}
{{- default (printf "%s-agent" (include "opvic.fullname" .)) .Values.agent.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.agent.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Control Plane URL
*/}}
{{- define "opvic.agent.controlPlaneURL" -}}
{{- if and (not .Values.agent.controlPlaneURL) (.Values.controlplane.enabled) }}
{{- default (printf "http://%s-control-plane" (include "opvic.fullname" .)) .Values.agent.controlPlaneURL }}
{{- else }}
{{- .Values.agent.controlPlaneURL }}
{{- end }}
{{- end }}
