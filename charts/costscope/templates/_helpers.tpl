{{- define "costscope.name" -}}
{{- .Chart.Name -}}
{{- end -}}

{{- define "costscope.fullname" -}}
{{- printf "%s-%s" .Release.Name .Chart.Name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
