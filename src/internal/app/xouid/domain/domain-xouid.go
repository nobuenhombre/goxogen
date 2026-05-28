package domainapp

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/google/uuid"
	pgxdb "github.com/nobuenhombre/suikat/pkg/db/connectors/postgres-pgx-db"
	"github.com/nobuenhombre/suikat/pkg/fico"
	"github.com/nobuenhombre/suikat/pkg/futi"
	"github.com/nobuenhombre/suikat/pkg/ge"

	"goxogen/src/internal/app/xouid/cli"
)

var sqlParameters = regexp.MustCompile(`(%%(\s|\S)*?%%)`)

// DomainService is the business-logic orchestrator for xouid.
type DomainService interface {
	Run() error
}

// AppDomain implements DomainService by orchestrating the full pipeline:
// validate query → check templates → parse params → explain in PG → generate code → write file.
type AppDomain struct {
	cliConfig cli.Service
	db        pgxdb.DBQuery
	queryText string
}

// New creates a new xouid domain service.
func New(cliConfig cli.Service, db pgxdb.DBQuery) (DomainService, error) {
	queryFile := fico.TxtFile(cliConfig.GetQuery())
	queryStr, err := queryFile.Read()
	if err != nil {
		return nil, ge.Pin(err)
	}

	return &AppDomain{
		cliConfig: cliConfig,
		db:        db,
		queryText: queryStr,
	}, nil
}

// Run executes the full xouid generation pipeline.
func (d *AppDomain) Run() error {
	if err := d.CheckQuery(); err != nil {
		return err
	}

	if err := d.CheckTemplatesExists(); err != nil {
		return err
	}

	qp, err := d.GetQueryParams()
	if err != nil {
		return err
	}

	plan, err := d.CheckExplainSQLInPostgresql(qp)
	if err != nil {
		return err
	}

	if d.cliConfig.GetVerbose() {
		log.Printf("EXPLAIN plan: %v", plan)
	}

	queryStr, err := d.CreateFuncQuery(qp)
	if err != nil {
		return err
	}

	return d.WritePackageFile(queryStr)
}

// CheckQuery validates that the query is UPDATE/INSERT/DELETE.
func (d *AppDomain) CheckQuery() error {
	queryLow := strings.ToLower(d.queryText)
	allowed := []string{SQLUpdate, SQLInsert, SQLDelete}
	found := false

	for _, word := range allowed {
		found = strings.Contains(queryLow, word)
		if found {
			break
		}
	}

	if !found {
		return &UnknownSQLConstructionError{
			Query: d.queryText,
		}
	}

	return nil
}

// CheckTemplatesExists verifies that required template files exist.
func (d *AppDomain) CheckTemplatesExists() error {
	templates := []string{TemplateNewPackage, TemplateQuery}

	for _, templateStr := range templates {
		if !futi.FileExists(filepath.Join(d.cliConfig.GetTemplatePath(), templateStr)) {
			return &TemplateNotFoundError{
				Template: templateStr,
			}
		}
	}

	return nil
}

// GetQueryParams parses %%param type%% parameters from the query text.
func (d *AppDomain) GetQueryParams() (*[]QueryParam, error) {
	matches := sqlParameters.FindAllString(d.queryText, -1)

	qp := make([]QueryParam, 0, len(matches))

	for _, paramStr := range matches {
		cleanParamStr := strings.Trim(paramStr, "%")
		paramParts := strings.Split(cleanParamStr, " ")

		if len(paramParts) != 2 {
			return nil, &IncorrectQueryParamError{
				Param: cleanParamStr,
			}
		}

		param := QueryParam{
			Name: paramParts[0],
			Type: paramParts[1],
		}

		qp = append(qp, param)
	}

	return &qp, nil
}

// CreateSQLQueryNormal replaces %%param type%% descriptors with $1, $2, ... placeholders.
func (d *AppDomain) CreateSQLQueryNormal(qp *[]QueryParam) string {
	normalSQL := d.queryText
	for i, p := range *qp {
		descriptor := p.GetDescriptor()
		normalSQL = strings.ReplaceAll(normalSQL, descriptor, fmt.Sprintf("$%v", i+1))
	}

	return normalSQL
}

// CreateSQLQueryExplain replaces %%param type%% descriptors with concrete example values.
func (d *AppDomain) CreateSQLQueryExplain(qp *[]QueryParam) (string, error) {
	explainSQL := fmt.Sprintf("EXPLAIN %v", d.queryText)
	for _, p := range *qp {
		descriptor := p.GetDescriptor()
		switch p.Type {
		case "int", "int32", "int64":
			explainSQL = strings.ReplaceAll(explainSQL, descriptor, "1")

		case "float", "float32", "float64":
			explainSQL = strings.ReplaceAll(explainSQL, descriptor, "1.25")

		case "string":
			explainSQL = strings.ReplaceAll(explainSQL, descriptor, "'hello'")

		case "bool":
			explainSQL = strings.ReplaceAll(explainSQL, descriptor, "true")

		case "uuid.UUID":
			newUUID := uuid.New()
			explainSQL = strings.ReplaceAll(explainSQL, descriptor, fmt.Sprintf("'%v'::uuid", newUUID.String()))

		case "time.Time":
			explainSQL = strings.ReplaceAll(explainSQL, descriptor, "'2020-02-03 15:45:10'")

		case "[]int", "[]int32", "[]int64":
			explainSQL = strings.ReplaceAll(explainSQL, descriptor, "'{1,2,3}'::int[]")

		default:
			return "", &UnknownQueryParamTypeError{
				Type: p.Type,
			}
		}
	}

	return explainSQL, nil
}

// CheckExplainSQLInPostgresql runs EXPLAIN on the generated SQL against PostgreSQL.
func (d *AppDomain) CheckExplainSQLInPostgresql(qp *[]QueryParam) (*[]string, error) {
	ctx := context.Background()

	explainSQL, err := d.CreateSQLQueryExplain(qp)
	if err != nil {
		return nil, err
	}

	q, err := d.db.Query(ctx, explainSQL)
	if err != nil {
		return nil, ge.Pin(err, ge.Params{"explainSQL": explainSQL})
	}
	defer q.Close()

	plan := make([]string, 0)
	for q.Next() {
		r := ""
		err = q.Scan(&r)
		if err != nil {
			return nil, err
		}
		plan = append(plan, r)
	}

	return &plan, nil
}

// CreateFuncQuery renders the query function template with the parsed data.
func (d *AppDomain) CreateFuncQuery(qp *[]QueryParam) (string, error) {
	normalSQL := d.CreateSQLQueryNormal(qp)

	t, err := template.ParseFiles(filepath.Join(d.cliConfig.GetTemplatePath(), TemplateQuery))
	if err != nil {
		return "", ge.Pin(err)
	}

	queryTemplateData := QueryTemplateData{
		Type:        d.cliConfig.GetQueryType(),
		Name:        d.cliConfig.GetQueryFunc(),
		QueryParams: *qp,
		SqlQuery:    normalSQL,
	}

	buf := new(bytes.Buffer)
	err = t.ExecuteTemplate(buf, TemplateNameQuery, queryTemplateData)
	if err != nil {
		return "", ge.Pin(err)
	}

	return buf.String(), nil
}

// CreateNewPackage renders the package header template.
func (d *AppDomain) CreateNewPackage() (string, error) {
	t, err := template.ParseFiles(filepath.Join(d.cliConfig.GetTemplatePath(), TemplateNewPackage))
	if err != nil {
		return "", ge.Pin(err)
	}

	packageTemplateData := PackageTemplateData{
		Package: d.cliConfig.GetPackage(),
		Schema:  d.cliConfig.GetSchema(),
	}

	buf := new(bytes.Buffer)
	err = t.ExecuteTemplate(buf, TemplateNameNewPackage, packageTemplateData)
	if err != nil {
		return "", ge.Pin(err)
	}

	return buf.String(), nil
}

// WritePackageFile creates or appends to the .xouid.go output file.
func (d *AppDomain) WritePackageFile(queryStr string) (err error) {
	packageFileName := fmt.Sprintf("%v%v.xouid.go", d.cliConfig.GetOut(), strings.ToLower(d.cliConfig.GetQueryType()))

	packageFile := fico.TxtFile(packageFileName)
	packageStr := ""

	if !futi.FileExists(packageFileName) {
		packageStr, err = d.CreateNewPackage()
		if err != nil {
			return err
		}

		err = packageFile.Write(packageStr)
		if err != nil {
			return err
		}
	}

	packageStr, err = packageFile.Read()
	if err != nil {
		return err
	}

	packageStr += queryStr

	err = packageFile.Write(packageStr)
	if err != nil {
		return err
	}

	return nil
}
