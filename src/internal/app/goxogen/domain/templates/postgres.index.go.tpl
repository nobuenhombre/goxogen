{{- $short := (shortname .Type.Name "err" "sqlstr" "db" "q" "res" "db.WriteLog" .Fields) -}}
{{- $shortResult := (print $short "Val") -}}
{{- $table := (schema .Schema .Type.Table.TableName) -}}
{{- $repoName := (print .Type.Name "Repository") -}}

// {{ .FuncName }} retrieves a row from '{{ $table }}' as a {{ .Type.Name }}.
//
// Generated from index '{{ .Index.IndexName }}'.
func Get{{ .FuncName }}(db pgxdb.DBQuery{{ goparamlist .Fields true true }}) ({{ if not .Index.IsUnique }}[]{{ end }}*{{ .Type.Name }}, error) {
	var err error

	start := time.Now()

	ctx := context.Background()

	// sql query
	// language=SQL
	const sqlstr = `
SELECT
	{{ colnames .Type.Fields }}
FROM
	{{ $table }}
WHERE
	{{ colnamesquery .Fields " AND " }}
`

	// run query
{{- if .Index.IsUnique }}
	var {{ $shortResult }} {{ .Type.Name }}
	{{- if .Type.PrimaryKey }}
	// @crud
	{{ $shortResult }}._exists = true
	// @end-crud
	{{ end }}

	err = db.QueryRow(ctx, sqlstr{{ goparamlist .Fields true false }}).Scan({{ fieldnames .Type.Fields (print "&" $shortResult) }})

	db.WriteLog(sqlstr, time.Since(start){{ goparamlist .Fields true false }})

	if err != nil {
		return nil, err
	}

	return &{{ $shortResult }}, nil
{{- else }}
	q, err := db.Query(ctx, sqlstr{{ goparamlist .Fields true false }})

	db.WriteLog(sqlstr, time.Since(start){{ goparamlist .Fields true false }})

	if err != nil {
		return nil, err
	}
	defer q.Close()

	// load results
	res := []*{{ .Type.Name }}{}
	for q.Next() {
		{{ $short }} := {{ .Type.Name }}{
		{{- if .Type.PrimaryKey }}
		// @crud
			_exists: true,
		// @end-crud
		{{ end -}}
		}

		// scan
		err = q.Scan({{ fieldnames .Type.Fields (print "&" $short) }})
		if err != nil {
			return nil, err
		}

		res = append(res, &{{ $short }})
	}

	return res, nil
{{- end }}
}

{{ if not .Index.IsUnique }}
// Get{{ .FuncName }}WithPagination retrieves a paginated set of rows from '{{ $table }}' as a {{ .Type.Name }}.
//
// Generated from index '{{ .Index.IndexName }}'.
func Get{{ .FuncName }}WithPagination(db pgxdb.DBQuery{{ goparamlist .Fields true true }}, limit, offset int) ([]*{{ .Type.Name }}, error) {
	var err error

	start := time.Now()

	ctx := context.Background()

	// sql query
	// language=SQL
	var sqlstr string
	{
		const baseSQL = `
SELECT
	{{ colnames .Type.Fields }}
FROM
	{{ $table }}
WHERE
	{{ colnamesquery .Fields " AND " }}
ORDER BY
	id ASC
`
		sqlstr = baseSQL + "\nLIMIT $" + fmt.Sprint(1{{ range .Fields }}+1{{ end }}) + " OFFSET $" + fmt.Sprint(2{{ range .Fields }}+1{{ end }})
	}

	// run query
	q, err := db.Query(ctx, sqlstr{{ goparamlist .Fields true false }}, limit, offset)

	db.WriteLog(sqlstr, time.Since(start){{ goparamlist .Fields true false }}, limit, offset)

	if err != nil {
		return nil, err
	}
	defer q.Close()

	// load results
	res := []*{{ .Type.Name }}{}
	for q.Next() {
		{{ $short }} := {{ .Type.Name }}{
		{{- if .Type.PrimaryKey }}
		// @crud
			_exists: true,
		// @end-crud
		{{ end -}}
		}

		// scan
		err = q.Scan({{ fieldnames .Type.Fields (print "&" $short) }})
		if err != nil {
			return nil, err
		}

		res = append(res, &{{ $short }})
	}

	return res, nil
}
{{ end }}

{{ if not .Index.IsUnique }}
// Get{{ .FuncName }}Count retrieves count of rows from '{{ $table }}' by index '{{ .Index.IndexName }}'.
func Get{{ .FuncName }}Count(db pgxdb.DBQuery{{ goparamlist .Fields true true }}) (int64, error) {
	var err error

	start := time.Now()

	ctx := context.Background()

	// sql query
	// language=SQL
	const sqlstr = `
SELECT
	COUNT(*)
FROM
	{{ $table }}
WHERE
	{{ colnamesquery .Fields " AND " }}
`

	// run query
	var count int64
	err = db.QueryRow(ctx, sqlstr{{ goparamlist .Fields true false }}).Scan(&count)

	db.WriteLog(sqlstr, time.Since(start){{ goparamlist .Fields true false }})

	if err != nil {
		return 0, err
	}

	return count, nil
}
{{ end }}

// ----- Index Methods for {{ .Type.Name }} -----

// @repo-start
{{/* Уникальные индексы */}}
{{- if .Index.IsUnique }}
    // Get{{ .FuncName }} возвращает одну запись по индексу '{{ .Index.IndexName }}'.
    func (repo *{{ $repoName }}) Get{{ .FuncName }}({{ goparamlist .Fields false true }}) (*{{ .Type.Name }}, error) {
        return Get{{ .FuncName }}(repo.db{{ goparamlist .Fields true false }})
    }
{{- end }}

{{/* Неуникальные индексы */}}
{{- if not .Index.IsUnique }}
    // FindAll{{ .FuncName }} возвращает все записи по индексу '{{ .Index.IndexName }}'.
    func (repo *{{ $repoName }}) FindAll{{ .FuncName }}({{ goparamlist .Fields false true }}) ([]*{{ .Type.Name }}, error) {
        return Get{{ .FuncName }}(repo.db{{ goparamlist .Fields true false }})
    }

    // FindAll{{ .FuncName }}WithPagination возвращает записи по индексу с пагинацией
    func (repo *{{ $repoName }}) FindAll{{ .FuncName }}WithPagination({{ goparamlist .Fields false true }}, limit, offset int) ([]*{{ .Type.Name }}, error) {
        return Get{{ .FuncName }}WithPagination(repo.db{{ goparamlist .Fields true false }}, limit, offset)
    }

    // FindAll{{ .FuncName }}Count возвращает количество записей по индексу '{{ .Index.IndexName }}'.
    func (repo *{{ $repoName }}) FindAll{{ .FuncName }}Count({{ goparamlist .Fields false true }}) (int64, error) {
        return Get{{ .FuncName }}Count(repo.db{{ goparamlist .Fields true false }})
    }
{{- end }}
// @repo-end
