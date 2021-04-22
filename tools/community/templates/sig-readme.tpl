{{- template "header" }}
# {{.Name}}
{{ .Description }}
{{ if .Leads }}
### Leads
The leads run operations and processes governing each group.
{{ range .Leads }}
- {{.Name}} (**[@{{.GitHub}}](https://github.com/{{.GitHub}})**)
{{- end }}
{{- end }}

{{- if .Subprojects }}
## Subprojects

The following subprojects are owned by {{ .Label }}:

{{- range .Subprojects }}
- **{{.Name}}**
{{- if .Description }}
  - Description: {{ .Description }}
{{- end }}
{{- if .Owners }}
  - Owner:
{{- range .Owners }}
    - {{ . }}
{{- end }}
{{- end }}
{{- end }}
{{- end }}

{{- if .Roadmaps }}

## Roadmaps
{{ range .Roadmaps }}
### {{ .Name }}
{{- range .Subprojects }}
- {{ .Name }}
  {{- range .Tasks }}
  - [ ] {{ .Name }}
  {{- end }}
{{- end }}
{{- end }}
{{- end }}