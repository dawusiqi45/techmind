{{/*
TechMind Helm Chart 辅助模板
*/}}

{{/* 生成全名 */}}
{{- define "techmind.fullname" -}}
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

{{/* 生成 Chart 名与版本标签 */}}
{{- define "techmind.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/* 通用标签选择器 */}}
{{- define "techmind.selectorLabels" -}}
app.kubernetes.io/name: {{ include "techmind.fullname" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/* 通用标签 */}}
{{- define "techmind.labels" -}}
helm.sh/chart: {{ include "techmind.chart" . }}
{{ include "techmind.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/* Server 标签选择器 */}}
{{- define "techmind.server.selectorLabels" -}}
app.kubernetes.io/name: {{ include "techmind.fullname" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/component: server
{{- end }}

{{/* Worker 标签选择器 */}}
{{- define "techmind.worker.selectorLabels" -}}
app.kubernetes.io/name: {{ include "techmind.fullname" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/component: worker
{{- end }}
