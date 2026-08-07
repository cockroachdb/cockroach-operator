{{- define "cockroach-operator.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "cockroach-operator.fullname" -}}
{{- if .Values.fullnameOverride }}{{ .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}{{ else }}{{ printf "%s-%s" .Release.Name (include "cockroach-operator.name" .) | trunc 63 | trimSuffix "-" }}{{ end }}
{{- end }}

{{- define "cockroach-operator.labels" -}}
app.kubernetes.io/name: {{ include "cockroach-operator.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{- define "cockroach-operator.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}{{ default (include "cockroach-operator.fullname" .) .Values.serviceAccount.name }}{{ else }}{{ required "serviceAccount.name is required when serviceAccount.create is false" .Values.serviceAccount.name }}{{ end }}
{{- end }}