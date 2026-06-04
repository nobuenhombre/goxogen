{{- $short := (shortname .Name "err" "res" "sqlstr" "db" "db.WriteLog") -}}
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
{{- if .PrimaryKey }}

	// xo fields
	_exists, _deleted bool
{{ end }}
}

{{ if .PrimaryKey }}
// Exists determines if the {{ .Name }} exists in the database.
func ({{ $short }} *{{ .Name }}) Exists() bool {
	return {{ $short }}._exists
}

// SetExists determines if the {{ .Name }} exists in the database.
func ({{ $short }} *{{ .Name }}) SetExists(exists bool) {
	{{ $short }}._exists = exists
}

// Deleted provides information if the {{ .Name }} has been deleted from the database.
func ({{ $short }} *{{ .Name }}) Deleted() bool {
	return {{ $short }}._deleted
}

// Insert inserts the {{ .Name }} to the database.
func ({{ $short }} *{{ .Name }}) Insert(db pgxdb.DBQuery) error {
	var err error

	start := time.Now()

	ctx := context.Background()

	// if already exist, bail
	if {{ $short }}._exists {
		return errors.New("insert failed: already exists")
	}

{{ if .Table.ManualPk }}
	// sql insert query, primary key must be provided
	// language=SQL
	const sqlstr = `
INSERT INTO {{ $table }} (
{{ colnames .Fields }}
) VALUES (
{{ colvals .Fields }}
) RETURNING {{ colname .PrimaryKey.Col }}
`

	// run query
	err = db.QueryRow(ctx, sqlstr, {{ fieldnames .Fields $short }}).Scan(&{{ $short }}.{{ .PrimaryKey.Name }})

	db.WriteLog(sqlstr, time.Since(start), {{ fieldnames .Fields $short }})

	if err != nil {
		return err
	}
{{ else }}
	// sql insert query, primary key provided by sequence
	// language=SQL
	const sqlstr = `
INSERT INTO {{ $table }} (
{{ colnames .Fields .PrimaryKey.Name }}
) VALUES (
{{ colvals .Fields .PrimaryKey.Name }}
) RETURNING {{ colname .PrimaryKey.Col }}
`

	// run query
	err = db.QueryRow(ctx, sqlstr, {{ fieldnames .Fields $short .PrimaryKey.Name }}).Scan(&{{ $short }}.{{ .PrimaryKey.Name }})

	db.WriteLog(sqlstr, time.Since(start), {{ fieldnames .Fields $short .PrimaryKey.Name }})

	if err != nil {
		return err
	}
{{ end }}

	// set existence
	{{ $short }}._exists = true

	return nil
}

{{ if ne (fieldnamesmulti .Fields $short .PrimaryKeyFields) "" }}
	// Update updates the {{ .Name }} in the database.
	func ({{ $short }} *{{ .Name }}) Update(db pgxdb.DBQuery) error {
		var err error

		start := time.Now()

		ctx := context.Background()

		// if doesn't exist, bail
		if !{{ $short }}._exists {
			return errors.New("update failed: does not exist")
		}

		// if deleted, bail
		if {{ $short }}._deleted {
			return errors.New("update failed: marked for deletion")
		}

		{{ if gt ( len .PrimaryKeyFields ) 1 }}
			// sql query with composite primary key
			// language=SQL
			const sqlstr = `
UPDATE {{ $table }} SET (
{{ colnamesmulti .Fields .PrimaryKeyFields }}
) = (
{{ colvalsmulti .Fields .PrimaryKeyFields }}
) WHERE {{ colnamesquerymulti .PrimaryKeyFields " AND " (getstartcount .Fields .PrimaryKeyFields) nil }}
`

			// run query
			_, err = db.Exec(ctx, sqlstr, {{ fieldnamesmulti .Fields $short .PrimaryKeyFields }}, {{ fieldnames .PrimaryKeyFields $short}})

			db.WriteLog(sqlstr, time.Since(start), {{ fieldnamesmulti .Fields $short .PrimaryKeyFields }}, {{ fieldnames .PrimaryKeyFields $short}})

			return err
		{{- else }}
			// sql query
			// language=SQL
			const sqlstr = `
UPDATE {{ $table }} SET (
{{ colnames .Fields .PrimaryKey.Name }}
) = (
{{ colvals .Fields .PrimaryKey.Name }}
) WHERE {{ colname .PrimaryKey.Col }} = ${{ colcount .Fields .PrimaryKey.Name }}
`

			// run query
			_, err = db.Exec(ctx, sqlstr, {{ fieldnames .Fields $short .PrimaryKey.Name }}, {{ $short }}.{{ .PrimaryKey.Name }})

			db.WriteLog(sqlstr, time.Since(start), {{ fieldnames .Fields $short .PrimaryKey.Name }}, {{ $short }}.{{ .PrimaryKey.Name }})

			return err
		{{- end }}
	}

	// Save saves the {{ .Name }} to the database.
	func ({{ $short }} *{{ .Name }}) Save(db pgxdb.DBQuery) error {
		if {{ $short }}.Exists() {
			return {{ $short }}.Update(db)
		}

		return {{ $short }}.Insert(db)
	}

	// @repo-start
    // Save saves the {{ .Name }} to the database.
    func (repo *{{ .Name }}Repository) Save({{ $short }} *{{ .Name }}) error {
        return {{ $short }}.Save(repo.db)
    }
	// @repo-end

	// Upsert performs an upsert for {{ .Name }}.
	//
	// NOTE: PostgreSQL 9.5+ only
	func ({{ $short }} *{{ .Name }}) Upsert(db pgxdb.DBQuery) error {
		var err error

		start := time.Now()

		ctx := context.Background()

		// if already exist, bail
		if {{ $short }}._exists {
			return errors.New("insert failed: already exists")
		}

		// sql query
		// language=SQL
		const sqlstr = `
INSERT INTO {{ $table }} (
{{ colnames .Fields }}
) VALUES (
{{ colvals .Fields }}
) ON CONFLICT ({{ colnames .PrimaryKeyFields }}) DO UPDATE SET (
{{ colnames .Fields }}
) = (
{{ colprefixnames .Fields "EXCLUDED" }}
)
`

		// run query
		_, err = db.Exec(ctx, sqlstr, {{ fieldnames .Fields $short }})

		db.WriteLog(sqlstr, time.Since(start), {{ fieldnames .Fields $short }})

		if err != nil {
			return err
		}

		// set existence
		{{ $short }}._exists = true

		return nil
}
{{ else }}
	// Update statements omitted due to lack of fields other than primary key
{{ end }}

// Delete deletes the {{ .Name }} from the database.
func ({{ $short }} *{{ .Name }}) Delete(db pgxdb.DBQuery) error {
	var err error

	start := time.Now()

	ctx := context.Background()

	// if doesn't exist, bail
	if !{{ $short }}._exists {
		return nil
	}

	// if deleted, bail
	if {{ $short }}._deleted {
		return nil
	}

	{{ if gt ( len .PrimaryKeyFields ) 1 }}
		// sql query with composite primary key
		// language=SQL
		const sqlstr = `
DELETE FROM {{ $table }}
WHERE {{ colnamesquery .PrimaryKeyFields " AND " }}
`

		// run query
		_, err = db.Exec(ctx, sqlstr, {{ fieldnames .PrimaryKeyFields $short }})

		db.WriteLog(sqlstr, time.Since(start), {{ fieldnames .PrimaryKeyFields $short }})

		if err != nil {
			return err
		}
	{{- else }}
		// sql query
		// language=SQL
		const sqlstr = `
DELETE FROM {{ $table }}
WHERE {{ colname .PrimaryKey.Col }} = $1
`

		// run query
		_, err = db.Exec(ctx, sqlstr, {{ $short }}.{{ .PrimaryKey.Name }})

		db.WriteLog(sqlstr, time.Since(start), {{ $short }}.{{ .PrimaryKey.Name }})

		if err != nil {
			return err
		}
	{{- end }}

	// set deleted
	{{ $short }}._deleted = true

	return nil
}

// @repo-start
// Delete deletes the {{ .Name }} from the database.
func (repo *{{ .Name }}Repository) Delete({{ $short }} *{{ .Name }}) error {
    return {{ $short }}.Delete(repo.db)
}
// @repo-end

{{- end }}

// GetAll{{ .Name }} returns all rows from '{{ .Schema }}.{{ .Table.TableName }}',
func GetAll{{ .Name }}(db pgxdb.DBQuery) ([]*{{ .Name }}, error) {
	ctx := context.Background()

	start := time.Now()

	// language=SQL
    const sqlstr = `
SELECT
{{ colnames .Fields }}
FROM {{ $table }}
ORDER BY
	id ASC
`

    q, err := db.Query(ctx, sqlstr)

	db.WriteLog(sqlstr, time.Since(start))

    if err != nil {
        return nil, err
    }
    defer q.Close()

    // load results
    var res []*{{ .Name }}
    for q.Next() {
        {{ $short }} := {{ .Name }}{}

        // scan
        err = q.Scan({{ fieldnames .Fields (print "&" $short) }})
        if err != nil {
            return nil, err
        }

        {{- if .PrimaryKey }}
        {{ $short }}.SetExists(true)
        {{- end }}

        res = append(res, &{{ $short }})
    }

    return res, nil
}

// GetAll{{ .Name }}WithPagination returns a paginated set of rows from '{{ .Schema }}.{{ .Table.TableName }}',
func GetAll{{ .Name }}WithPagination(db pgxdb.DBQuery, limit, offset int) ([]*{{ .Name }}, error) {
	ctx := context.Background()

	start := time.Now()

	// language=SQL
    const sqlstr = `
SELECT
{{ colnames .Fields }}
FROM {{ $table }}
ORDER BY
	id ASC
LIMIT $1 OFFSET $2
`

    q, err := db.Query(ctx, sqlstr, limit, offset)

	db.WriteLog(sqlstr, time.Since(start), limit, offset)

    if err != nil {
        return nil, err
    }
    defer q.Close()

    // load results
    var res []*{{ .Name }}
    for q.Next() {
        {{ $short }} := {{ .Name }}{}

        // scan
        err = q.Scan({{ fieldnames .Fields (print "&" $short) }})
        if err != nil {
            return nil, err
        }

        {{- if .PrimaryKey }}
        {{ $short }}.SetExists(true)
        {{- end }}

        res = append(res, &{{ $short }})
    }

    return res, nil
}

// GetAll{{ .Name }}Count returns count of all rows from '{{ .Schema }}.{{ .Table.TableName }}',
func GetAll{{ .Name }}Count(db pgxdb.DBQuery) (int64, error) {
	ctx := context.Background()

	start := time.Now()

	// language=SQL
	const sqlstr = `SELECT COUNT(*) FROM {{ $table }}`

	var count int64
	err := db.QueryRow(ctx, sqlstr).Scan(&count)

	db.WriteLog(sqlstr, time.Since(start))

	if err != nil {
		return 0, err
	}

	return count, nil
}

// Get{{ .Name }}sBySQL returns rows from '{{ .Schema }}.{{ .Table.TableName }}' by your SQL,
func Get{{ .Name }}sBySQL(db pgxdb.DBQuery, sqlstr string, args ...interface{}) ([]*{{ .Name }}, error) {
	ctx := context.Background()

	start := time.Now()

    q, err := db.Query(ctx, sqlstr, args...)

	db.WriteLog(sqlstr, time.Since(start), args...)

    if err != nil {
        return nil, err
    }
    defer q.Close()

    // load results
    var res []*{{ .Name }}
    for q.Next() {
        {{ $short }} := {{ .Name }}{}

        // scan
        err = q.Scan({{ fieldnames .Fields (print "&" $short) }})
        if err != nil {
            return nil, err
        }

        res = append(res, &{{ $short }})
    }

    return res, nil
}

// Get{{ .Name }}sBySQLWithPagination returns a paginated set of rows from '{{ .Schema }}.{{ .Table.TableName }}' by your SQL,
func Get{{ .Name }}sBySQLWithPagination(db pgxdb.DBQuery, sqlstr string, limit, offset int, args ...interface{}) ([]*{{ .Name }}, error) {
	ctx := context.Background()

	start := time.Now()

    paginatedSQL := sqlstr + fmt.Sprintf("\nLIMIT $%d OFFSET $%d", len(args)+1, len(args)+2)
    paginatedArgs := append(args, limit, offset)

    q, err := db.Query(ctx, paginatedSQL, paginatedArgs...)

	db.WriteLog(paginatedSQL, time.Since(start), paginatedArgs...)

    if err != nil {
        return nil, err
    }
    defer q.Close()

    // load results
    var res []*{{ .Name }}
    for q.Next() {
        {{ $short }} := {{ .Name }}{}

        // scan
        err = q.Scan({{ fieldnames .Fields (print "&" $short) }})
        if err != nil {
            return nil, err
        }

        res = append(res, &{{ $short }})
    }

    return res, nil
}

// Get{{ .Name }}sBySQLCount returns count of rows from '{{ .Schema }}.{{ .Table.TableName }}' by your SQL,
func Get{{ .Name }}sBySQLCount(db pgxdb.DBQuery, sqlstr string, args ...interface{}) (int64, error) {
	ctx := context.Background()

	start := time.Now()

	countSQL := `SELECT COUNT(*) FROM (` + sqlstr + `) AS count_query`

	var count int64
	err := db.QueryRow(ctx, countSQL, args...).Scan(&count)

	db.WriteLog(countSQL, time.Since(start), args...)

	if err != nil {
		return 0, err
	}

	return count, nil
}

// GetLast{{ .Name }} returns last row from '{{ .Schema }}.{{ .Table.TableName }}',
//
// errors possible
//  1. pgx.ErrNoRows
//  2. pgconn.PgError
func GetLastID{{ .Name }}(db pgxdb.DBQuery) (*{{ .Name }}, error) {
    ctx := context.Background()

    start := time.Now()

    // language=SQL
    const sqlstr = `
SELECT
{{ colnames .Fields }}
FROM {{ $table }}
ORDER BY
id DESC
LIMIT 1
`

    var {{ $short }} {{ .Name }}
    err := db.QueryRow(ctx, sqlstr).Scan({{ fieldnames .Fields (print "&" $short) }})

    db.WriteLog(sqlstr, time.Since(start))

    if err != nil {
        return nil, err
    }

    {{- if .PrimaryKey }}
    {{ $short }}._exists = true
    {{ $short }}._deleted = false
    {{- end }}

    return &{{ $short }}, nil
}

// @repo-start
{{/* ===== Репозиторий ===== */}}
// {{ .Name }}Repository реализует работу с таблицей '{{ .Table.TableName }}'.
type {{ .Name }}Repository struct {
    db pgxdb.DBQuery
}

// New{{ .Name }}Repository создает новый репозиторий.
func New{{ .Name }}Repository(db pgxdb.DBQuery) *{{ .Name }}Repository {
    return &{{ .Name }}Repository{db: db}
}

// GetAll возвращает все записи
func (repo *{{ .Name }}Repository) GetAll() ([]*{{ .Name }}, error) {
    return GetAll{{ .Name }}(repo.db)
}

// GetAllWithPagination возвращает записи с пагинацией
func (repo *{{ .Name }}Repository) GetAllWithPagination(limit, offset int) ([]*{{ .Name }}, error) {
    return GetAll{{ .Name }}WithPagination(repo.db, limit, offset)
}

// GetAllCount возвращает количество записей
func (repo *{{ .Name }}Repository) GetAllCount() (int64, error) {
    return GetAll{{ .Name }}Count(repo.db)
}

// GetBySQL возвращает записи по произвольному SQL
func (repo *{{ .Name }}Repository) GetBySQL(sqlstr string, args ...interface{}) ([]*{{ .Name }}, error) {
    return Get{{ .Name }}sBySQL(repo.db, sqlstr, args...)
}

// GetBySQLWithPagination возвращает записи по произвольному SQL с пагинацией
func (repo *{{ .Name }}Repository) GetBySQLWithPagination(sqlstr string, limit, offset int, args ...interface{}) ([]*{{ .Name }}, error) {
    return Get{{ .Name }}sBySQLWithPagination(repo.db, sqlstr, limit, offset, args...)
}

// GetBySQLCount возвращает количество записей по произвольному SQL
func (repo *{{ .Name }}Repository) GetBySQLCount(sqlstr string, args ...interface{}) (int64, error) {
    return Get{{ .Name }}sBySQLCount(repo.db, sqlstr, args...)
}

// GetLastID возвращает последний ID
func (repo *{{ .Name }}Repository) GetLastID() (*{{ .Name }}, error) {
return GetLastID{{ .Name }}(repo.db)
}
// @repo-end
