{{- define "gateway.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name "gateway" | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}

{{- define "gateway.labels" -}}
helm.sh/chart: {{ printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
app.kubernetes.io/name: gateway
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/part-of: habr
{{- end }}

{{- define "gateway.selectorLabels" -}}
app.kubernetes.io/name: gateway
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}
