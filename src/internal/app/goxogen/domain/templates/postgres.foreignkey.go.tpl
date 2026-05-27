{{- $short := (shortname .Type.Name) -}}
// {{ .Name }} returns the {{ .RefType.Name }} associated with the {{ .Type.Name }}'s {{ .Field.Name }} ({{ .Field.Col.ColumnName }}).
//
// Generated from foreign key '{{ .ForeignKey.ForeignKeyName }}'.
func ({{ $short }} *{{ .Type.Name }}) Get{{ .Name }}(db pgxdb.DBQuery) (*{{ .RefType.Name }}, error) {
	return Get{{ .RefType.Name }}By{{ .RefField.Name }}(db, {{ convext $short .Field .RefField }})
}

