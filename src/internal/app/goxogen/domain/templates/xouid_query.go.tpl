{{define "xouidquery"}}

{{- $repoName := (print .Type "Repository") -}}

// Custom Query
func {{ .Name }}(db pgxdb.DBQuery{{ range .QueryParams }}, {{ .Name }} {{ .Type }}{{ end }}) (err error) {
    ctx := context.Background()

    start := time.Now()

	// language=SQL
    sqlstr := `
{{ .SqlQuery }}
	`

    _, err = db.Exec(ctx, sqlstr{{ range .QueryParams }}, {{ .Name }}{{ end }})

    db.WriteLog(sqlstr, time.Since(start){{ range .QueryParams }}, {{ .Name }}{{ end }})

    return err
}

// @repo-start
func (repo *{{ $repoName }}) {{ .Name }}({{ range $i, $p := .QueryParams }}{{ if $i }}, {{ end }}{{ $p.Name }} {{ $p.Type }}{{ end }}) (error) {
    return {{ .Name }}(repo.db{{ range .QueryParams }}, {{ .Name }}{{ end }})
}
// @repo-end

{{end}}
