{{/*
Expand the name of the chart.
*/}}
{{- define "kubenexus.name" -}}
{{- default .Chart.Name .Values.global.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "kubenexus.fullname" -}}
{{- if .Values.global.fullnameOverride }}
{{- .Values.global.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.global.nameOverride }}
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
{{- define "kubenexus.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "kubenexus.labels" -}}
helm.sh/chart: {{ include "kubenexus.chart" . }}
{{ include "kubenexus.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "kubenexus.selectorLabels" -}}
app.kubernetes.io/name: {{ include "kubenexus.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Service Account Name
*/}}
{{- define "kubenexus.serviceAccountName" -}}
{{- if .Values.rbac.create }}
{{- default (printf "%s-sa" (include "kubenexus.fullname" .)) .Values.rbac.serviceAccountName }}
{{- else }}
{{- default "default" .Values.rbac.serviceAccountName }}
{{- end }}
{{- end }}

{{/*
Database Host resolution (Built-in or External)
*/}}
{{- define "kubenexus.dbHost" -}}
{{- if .Values.database.host }}
{{- .Values.database.host }}
{{- else if .Values.postgresql.enabled }}
{{- printf "%s-postgres" (include "kubenexus.fullname" .) }}
{{- else }}
{{- "localhost" }}
{{- end }}
{{- end }}

{{/*
Redis Host resolution (Built-in or External)
*/}}
{{- define "kubenexus.redisHost" -}}
{{- if .Values.redis.host }}
{{- .Values.redis.host }}
{{- else if .Values.redisInternal.enabled }}
{{- printf "%s-redis" (include "kubenexus.fullname" .) }}
{{- else }}
{{- "localhost" }}
{{- end }}
{{- end }}
