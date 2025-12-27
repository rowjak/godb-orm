package generator

import (
	"bytes"
	"text/template"
)

// TemplateData holds all data needed for struct template rendering
type TemplateData struct {
	PackageName string
	Imports     string
	StructName  string
	TableName   string
	Fields      []StructField
	HasTime     bool
	HasJSON     bool
	HasUUID     bool
}

// StructTemplate is the template for generating Go struct files
const StructTemplate = `package {{.PackageName}}
{{if .Imports}}

{{.Imports}}
{{end}}

// {{.StructName}} represents the {{.TableName}} table
type {{.StructName}} struct {
{{- range .Fields}}
	{{.Name}} {{.Type}} ` + "`{{.Tags}}`" + `{{if .Comment}} {{.Comment}}{{end}}
{{- end}}
}

// TableName returns the table name for GORM
func ({{.StructName}}) TableName() string {
	return "{{.TableName}}"
}
`

// TemplateRenderer handles template rendering
type TemplateRenderer struct {
	template *template.Template
}

// NewTemplateRenderer creates a new TemplateRenderer instance
func NewTemplateRenderer() (*TemplateRenderer, error) {
	tmpl, err := template.New("struct").Parse(StructTemplate)
	if err != nil {
		return nil, err
	}
	return &TemplateRenderer{template: tmpl}, nil
}

// Render renders the template with the given data
func (tr *TemplateRenderer) Render(data *TemplateData) (string, error) {
	var buf bytes.Buffer
	if err := tr.template.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// RenderBytes renders the template and returns bytes
func (tr *TemplateRenderer) RenderBytes(data *TemplateData) ([]byte, error) {
	var buf bytes.Buffer
	if err := tr.template.Execute(&buf, data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// BuildTemplateData creates TemplateData from GeneratedFile and detected imports
func BuildTemplateData(genFile *GeneratedFile, importMgr *ImportManager) *TemplateData {
	return &TemplateData{
		PackageName: genFile.PackageName,
		Imports:     genFile.Imports,
		StructName:  genFile.StructName,
		TableName:   genFile.TableName,
		Fields:      genFile.Fields,
		HasTime:     importMgr.Has(WellKnownImports.Time),
		HasJSON:     importMgr.Has(WellKnownImports.Datatypes),
		HasUUID:     importMgr.Has(WellKnownImports.UUID),
	}
}

// DetectRequiredImports scans fields and detects which imports are needed
// This implements the "smart import" feature
func DetectRequiredImports(fields []StructField) *ImportManager {
	importMgr := NewImportManager()

	for _, field := range fields {
		goType := field.Type

		// Check for time.Time
		if goType == "time.Time" || goType == "*time.Time" {
			importMgr.Add(WellKnownImports.Time)
		}

		// Check for datatypes.JSON
		if goType == "datatypes.JSON" || goType == "*datatypes.JSON" {
			importMgr.Add(WellKnownImports.Datatypes)
		}

		// Check for uuid.UUID
		if goType == "uuid.UUID" || goType == "*uuid.UUID" {
			importMgr.Add(WellKnownImports.UUID)
		}

		// Also add from ImportPath if specified
		if field.ImportPath != "" {
			importMgr.Add(field.ImportPath)
		}
	}

	return importMgr
}
