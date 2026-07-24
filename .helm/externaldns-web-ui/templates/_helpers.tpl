{{- define "externaldns-web-ui.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "externaldns-web-ui.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := include "externaldns-web-ui.name" . -}}
{{- if contains $name .Release.Name -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{- define "externaldns-web-ui.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" -}}
{{- end -}}

{{- define "externaldns-web-ui.labels" -}}
helm.sh/chart: {{ include "externaldns-web-ui.chart" . }}
app.kubernetes.io/name: {{ include "externaldns-web-ui.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/version: {{ .Values.global.appVersion | default .Chart.AppVersion | quote }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/part-of: externaldns-web-ui
{{- end -}}

{{- define "externaldns-web-ui.selectorLabels" -}}
app.kubernetes.io/name: {{ include "externaldns-web-ui.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{- define "externaldns-web-ui.backendImage" -}}
{{- $registry := .Values.global.imageRegistry -}}
{{- $repo := .Values.backend.image.repository -}}
{{- $tag := .Values.backend.image.tag | default .Values.global.appVersion | default "latest" -}}
{{- if $registry -}}
{{- printf "%s/%s:%s" $registry $repo $tag -}}
{{- else -}}
{{- printf "%s:%s" $repo $tag -}}
{{- end -}}
{{- end -}}

{{- define "externaldns-web-ui.frontendImage" -}}
{{- $registry := .Values.global.imageRegistry -}}
{{- $repo := .Values.frontend.image.repository -}}
{{- $tag := .Values.frontend.image.tag | default .Values.global.appVersion | default "latest" -}}
{{- if $registry -}}
{{- printf "%s/%s:%s" $registry $repo $tag -}}
{{- else -}}
{{- printf "%s:%s" $repo $tag -}}
{{- end -}}
{{- end -}}

{{- define "externaldns-web-ui.serviceAccountName" -}}
{{- if .Values.serviceAccount.create -}}
{{- default (include "externaldns-web-ui.fullname" .) .Values.serviceAccount.name -}}
{{- else -}}
{{- default "default" .Values.serviceAccount.name -}}
{{- end -}}
{{- end -}}

{{- define "externaldns-web-ui.backendFullname" -}}
{{- printf "%s-backend" (include "externaldns-web-ui.fullname" .) -}}
{{- end -}}

{{- define "externaldns-web-ui.frontendFullname" -}}
{{- printf "%s-frontend" (include "externaldns-web-ui.fullname" .) -}}
{{- end -}}

{{- define "externaldns-web-ui.frontendUrl" -}}
{{- if .Values.auth.frontendUrl -}}
{{- .Values.auth.frontendUrl -}}
{{- else if .Values.ingress.host -}}
{{- printf "https://%s" .Values.ingress.host -}}
{{- else -}}
{{- "" -}}
{{- end -}}
{{- end -}}
