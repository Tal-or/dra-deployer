{{/*
Expand the name of the chart.
*/}}
{{- define "dra-driver-memory.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "dra-driver-memory.fullname" -}}
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
{{- define "dra-driver-memory.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "dra-driver-memory.labels" -}}
helm.sh/chart: {{ include "dra-driver-memory.chart" . }}
{{ include "dra-driver-memory.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "dra-driver-memory.selectorLabels" -}}
app.kubernetes.io/name: {{ include "dra-driver-memory.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "dra-driver-memory.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (printf "%s-service-account" (include "dra-driver-memory.fullname" .)) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Create the name of the cluster role to use
*/}}
{{- define "dra-driver-memory.clusterRoleName" -}}
{{- printf "%s-role" (include "dra-driver-memory.fullname" .) }}
{{- end }}

{{/*
Create the name of the cluster role binding to use
*/}}
{{- define "dra-driver-memory.clusterRoleBindingName" -}}
{{- printf "%s-role-binding" (include "dra-driver-memory.fullname" .) }}
{{- end }}

