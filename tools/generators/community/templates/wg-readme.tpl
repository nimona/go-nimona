{{- template "header" }}
# {{.Name}} Working Group
{{ .Description }}
{{ if .Chairs }}
### Chairs
The Chairs of the WG run operations and processes governing the WG.
{{ range .Chairs }}
- {{.Name}} (**[@{{.GitHub}}](https://github.com/{{.GitHub}})**)
{{- end }}
{{- end }}

{{- if .Subprojects }}
## Subprojects

The following subprojects are owned by wg-{{.Label}}:

{{- range .Subprojects }}
- **{{.Name}}**
{{- if .Description }}
  - Description: {{ .Description }}
{{- end }}
{{- if .Owners }}
  - Owners:
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