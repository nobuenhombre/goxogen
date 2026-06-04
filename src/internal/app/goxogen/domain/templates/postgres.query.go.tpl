{{- $repoName := (print .Type.Name "Repository") -}}
{{ $short := (shortname .Type.Name "err" "sqlstr" "db" "q" "res" "db.WriteLog" .QueryParams) }}
{{ $queryComments := .QueryComments }}
{{ if .Comment }}
// {{ .Comment }}
{{ else }}
// {{ .Name }} runs a custom query, returning results as {{ .Type.Name }}.
{{ end -}}
func {{ .Name }} (db pgxdb.DBQuery{{ range .QueryParams }}, {{ .Name }} {{ .Type }}{{ end }}) ({{ if not .OnlyOne }}[]{{ end }}*{{ .Type.Name }}, error) {
	var err error

	start := time.Now()

	ctx := context.Background()

	// sql query
	{{ if .Interpolate }}var{{ else }}const{{ end }} sqlstr = {{ range $i, $l := .Query }}{{ if $i }} +"\n"+{{ end }}{{ if (index $queryComments $i) }} // {{ index $queryComments $i }}{{ end }}{{ if $i }}
	{{end -}}`{{ $l }}`{{ end }}

	// run query
{{- if .OnlyOne }}
	var {{ $short }} {{ .Type.Name }}
	err = db.QueryRow(ctx, sqlstr{{ range .QueryParams }}, {{ .Name }}{{ end }}).Scan({{ fieldnames .Type.Fields (print "&" $short) }})

	db.WriteLog(sqlstr, time.Since(start){{ range .QueryParams }}{{ if not .Interpolate }}, {{ .Name }}{{ end }}{{ end }})

	if err != nil {
		return nil, err
	}

	{{ $short }}._exists = true
	{{ $short }}._deleted = false

	return &{{ $short }}, nil
{{- else }}
	q, err := db.Query(ctx, sqlstr{{ range .QueryParams }}, {{ .Name }}{{ end }})

	db.WriteLog(sqlstr, time.Since(start){{ range .QueryParams }}{{ if not .Interpolate }}, {{ .Name }}{{ end }}{{ end }})

	if err != nil {
		return nil, err
	}
	defer q.Close()

	// load results
	res := []*{{ .Type.Name }}{}
	for q.Next() {
		{{ $short }} := {{ .Type.Name }}{}

		// scan
		err = q.Scan({{ fieldnames .Type.Fields (print "&" $short) }})
		if err != nil {
			return nil, err
		}

		{{ $short }}._exists = true
		{{ $short }}._deleted = false

		res = append(res, &{{ $short }})
	}

	return res, nil
{{- end }}
}

{{ if not .OnlyOne }}
// {{ if .Comment }}{{ .Comment }} with pagination{{ else }}{{ .Name }}WithPagination runs a custom query with pagination{{ end }}
func {{ .Name }}WithPagination (db pgxdb.DBQuery{{ range .QueryParams }}, {{ .Name }} {{ .Type }}{{ end }}, limit, offset int) ([]*{{ .Type.Name }}, error) {
	var err error

	start := time.Now()

	ctx := context.Background()

	// sql query
	{{ if .Interpolate }}var{{ else }}var{{ end }} baseSQL = {{ range $i, $l := .Query }}{{ if $i }} +"\n"+{{ end }}{{ if (index $queryComments $i) }} // {{ index $queryComments $i }}{{ end }}{{ if $i }}
	{{end -}}`{{ $l }}`{{ end }}

	sqlstr := baseSQL + "\nLIMIT $" + fmt.Sprint(1{{ range .QueryParams }}+1{{ end }}) + " OFFSET $" + fmt.Sprint(2{{ range .QueryParams }}+1{{ end }})

	// run query
	q, err := db.Query(ctx, sqlstr{{ range .QueryParams }}, {{ .Name }}{{ end }}, limit, offset)

	db.WriteLog(sqlstr, time.Since(start){{ range .QueryParams }}{{ if not .Interpolate }}, {{ .Name }}{{ end }}{{ end }}, limit, offset)

	if err != nil {
		return nil, err
	}
	defer q.Close()

	// load results
	res := []*{{ .Type.Name }}{}
	for q.Next() {
		{{ $short }} := {{ .Type.Name }}{}

		// scan
		err = q.Scan({{ fieldnames .Type.Fields (print "&" $short) }})
		if err != nil {
			return nil, err
		}

		{{ $short }}._exists = true
		{{ $short }}._deleted = false

		res = append(res, &{{ $short }})
	}

	return res, nil
}
{{ end }}

// @repo-start
{{ if .Comment }}
// {{ .Comment }}
{{ else }}
// {{ .Name }} runs a custom query, returning results as {{ .Type.Name }}.
{{ end -}}
func (repo *{{ $repoName }}) {{ .Name }}({{ range $i, $p := .QueryParams }}{{ if $i }}, {{ end }}{{ $p.Name }} {{ $p.Type }}{{ end }}) ({{ if not .OnlyOne }}[]{{ end }}*{{ .Type.Name }}, error) {
	return {{ .Name }}(repo.db{{ range .QueryParams }}, {{ .Name }}{{ end }})
}
{{ if not .OnlyOne }}
// {{ if .Comment }}{{ .Comment }} with pagination{{ else }}{{ .Name }}WithPagination runs a custom query with pagination from repository{{ end }}
func (repo *{{ $repoName }}) {{ .Name }}WithPagination({{ if .QueryParams }}{{ range $i, $p := .QueryParams }}{{ if $i }}, {{ end }}{{ $p.Name }} {{ $p.Type }}{{ end }}, {{ end }}limit, offset int) ([]*{{ .Type.Name }}, error) {
	return {{ .Name }}WithPagination(repo.db{{ range .QueryParams }}, {{ .Name }}{{ end }}, limit, offset)
}
{{ end }}
// @repo-end
