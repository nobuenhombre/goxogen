{{- $table := (schema .Schema .Table.TableName) -}}
{{- if .Comment -}}
// {{ .Comment }}
{{- else -}}
// {{ .Name }} represents a row from '{{ $table }}'.
{{- end }}
type {{ .Name }} struct {
{{- range .Fields }}
	{{ .Name }} {{ retype .Type }} `json:"{{ .Col.ColumnName }}"` // {{ .Col.ColumnName }}
{{- end }}

	// xo fields
	_exists, _deleted bool
}

func (data *{{ .Name }}) SetExists(exists bool) {
	data._exists = exists
}

// @repo-start
{{/* ===== Репозиторий ===== */}}
// {{ .Name }}Repository реализует работу с custom типом '{{ .Name }}'.
type {{ .Name }}Repository struct {
db pgxdb.DBQuery
}

// New{{ .Name }}Repository создает новый репозиторий.
func New{{ .Name }}Repository(db pgxdb.DBQuery) *{{ .Name }}Repository {
return &{{ .Name }}Repository{db: db}
}

// @repo-end
