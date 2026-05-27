{{- $notVoid := (ne .Proc.ReturnType "void") -}}
{{- $proc := (schema .Schema .Proc.ProcName) -}}
{{- if ne .Proc.ReturnType "trigger" -}}
// {{ .Name }} calls the stored procedure '{{ $proc }}({{ .ProcParams }}) {{ .Proc.ReturnType }}' on db.
func {{ .Name }}(db pgxdb.DBQuery{{ goparamlist .Params true true }}) ({{ if $notVoid }}{{ retype .Return.Type }}, {{ end }}error) {
	var err error

	start := time.Now()

	ctx := context.Background()

	// sql query
	const sqlstr = `SELECT {{ $proc }}({{ colvals .Params }})`

	// run query
{{- if $notVoid }}
	var ret {{ retype .Return.Type }}

	err = db.QueryRow(ctx, sqlstr{{ goparamlist .Params true false }}).Scan(&ret)

	db.WriteLog(sqlstr, time.Since(start){{ goparamlist .Params true false }})

	if err != nil {
		return {{ reniltype .Return.NilType }}, err
	}

	return ret, nil
{{- else }}
	_, err = db.Exec(ctx, sqlstr)

	db.WriteLog(sqlstr, time.Since(start))

	return err
{{- end }}
}
{{- end }}

