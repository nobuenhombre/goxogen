package domainapp

import "fmt"

// SQL query type constants.
const (
	SQLUpdate = "update"
	SQLInsert = "insert"
	SQLDelete = "delete"
)

// Template file name constants.
const (
	TemplateNewPackage     = "xouid_package.go.tpl"
	TemplateNameNewPackage = "xouidpackage"
	TemplateQuery          = "xouid_query.go.tpl"
	TemplateNameQuery      = "xouidquery"
)

// QueryParam represents a parsed parameter from a SQL query.
type QueryParam struct {
	Name string
	Type string
}

// GetDescriptor returns the full %%param type%% descriptor.
func (qp *QueryParam) GetDescriptor() string {
	return "%%" + qp.Name + " " + qp.Type + "%%"
}

// QueryTemplateData holds data for rendering the query function template.
type QueryTemplateData struct {
	Type        string
	Name        string
	QueryParams []QueryParam
	SqlQuery    string
}

// PackageTemplateData holds data for rendering the package file template.
type PackageTemplateData struct {
	Package string
	Schema  string
}

// UnknownSQLConstructionError indicates the query is not UPDATE/INSERT/DELETE.
type UnknownSQLConstructionError struct {
	Query string
}

func (e *UnknownSQLConstructionError) Error() string {
	return fmt.Sprintf("unknown sql construction [%v]", e.Query)
}

// TemplateNotFoundError indicates a required template file was not found.
type TemplateNotFoundError struct {
	Template string
}

func (e *TemplateNotFoundError) Error() string {
	return fmt.Sprintf("template file not found [%v]", e.Template)
}

// IncorrectQueryParamError indicates a query parameter has wrong format.
type IncorrectQueryParamError struct {
	Param string
}

func (e *IncorrectQueryParamError) Error() string {
	return fmt.Sprintf("incorrect query param [%v]", e.Param)
}

// UnknownQueryParamTypeError indicates an unsupported parameter type.
type UnknownQueryParamTypeError struct {
	Type string
}

func (e *UnknownQueryParamTypeError) Error() string {
	return fmt.Sprintf("unknown query param type [%v]", e.Type)
}

// QueryExplainResult holds the result of an EXPLAIN query.
type QueryExplainResult struct {
	Plan []string
}
