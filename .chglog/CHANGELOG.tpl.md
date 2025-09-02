# {{ .Info.Title }}

{{ if .Versions -}}
{{ range .Versions }}
## {{ if .Tag.Previous }}[{{ .Tag.Name }}]({{ $.Info.RepositoryURL }}/compare/{{ .Tag.Previous.Name }}...{{ .Tag.Name }}){{ else }}{{ .Tag.Name }}{{ end }} - {{ datetime "2006-01-02" .Tag.Date }}

{{ if .CommitGroups -}}
{{ range .CommitGroups -}}
### {{ .Title }}
{{ range .Commits -}}
- {{ if .Scope }}**{{ .Scope }}:** {{ end }}{{ .Subject }} ([{{ .Hash.Short }}]({{ $.Info.RepositoryURL }}/commit/{{ .Hash.Long }}))
{{ end }}
{{ end -}}
{{ end -}}

{{ if .RevertCommits -}}
### ⏪ Reverts
{{ range .RevertCommits -}}
- {{ .Revert.Header }} ([{{ .Hash.Short }}]({{ $.Info.RepositoryURL }}/commit/{{ .Hash.Long }}))
{{ end -}}
{{ end -}}

{{ if .NoteGroups -}}
{{ range .NoteGroups -}}
### 💥 {{ .Title }}
{{ range .Notes }}
- {{ .Body }}
{{ end }}
{{ end -}}
{{ end -}}

{{ end -}}
{{ else }}
*(No releases yet)*  
{{ end -}}
