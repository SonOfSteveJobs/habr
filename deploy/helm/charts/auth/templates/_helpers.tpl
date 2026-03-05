{{- define "auth.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name "auth" | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}

{{- define "auth.labels" -}}
helm.sh/chart: {{ printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
app.kubernetes.io/name: auth
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/part-of: habr
{{- end }}

{{- define "auth.selectorLabels" -}}
app.kubernetes.io/name: auth
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}
