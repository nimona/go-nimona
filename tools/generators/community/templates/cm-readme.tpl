{{- template "header" }}
# Community

## Working Groups
{{- range . }}

### [{{ .Name }}]({{ .Label }}/README.md)

{{ .Description }}

__Subprojects:__
{{- range .Subprojects }}
  - {{ .Name }}
{{- end }}

{{- end }}

