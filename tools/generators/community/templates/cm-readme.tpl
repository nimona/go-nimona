{{- template "header" }}
# Special Interest Groups
{{- range . }}

# [{{ .Name }}]({{ .Label }}/README.md)

{{ .Description }}

__Leads:__
{{- range .Leads }}
  - {{ .Name }}  (**[@{{ .GitHub }}](https://github.com/{{ .GitHub }})**)
{{- end }}

__Subprojects:__
{{- range .Subprojects }}
  - {{ .Name }}
{{- end }}

{{- end }}

