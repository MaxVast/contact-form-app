{{- define "contact-form-app.fullname" -}}
{{- .Release.Name -}}
{{- end -}}

{{- define "contact-form-app.labels" -}}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}