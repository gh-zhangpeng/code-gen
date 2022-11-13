package templates

// Model template for model
const Model = FileHeader + `
package {{.StructInfo.Package}}

import (
	"encoding/json"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

{{if .TableName -}}const TableName{{.StructName}} = "{{.TableName}}"{{- end}}
{{if .TableName -}}const TableCount{{.StructName}} = {{.TableCount}}{{- end}}

// {{.StructName}} {{.StructComment}}
type {{.StructName}} struct {
    {{range .Fields}}
	{{if .MultilineComment -}}
	/*
{{.Comment}}
    */
	{{end -}}
    {{.Name}} {{.DbType}} ` + "`{{.Tags}}` " +
	"{{if not .MultilineComment}}{{if .Comment}}// {{.Comment}}{{end}}{{end}}" +
	`{{end}}
}

{{if .TableName -}}
// TableName {{.StructName}}'s table name
func (*{{.StructName}}) TableName() string {
    return TableName{{.StructName}}
}
{{- end}}

{{if gt .TableCount 1}}
// TableNameOf{{.StructName}} {{.StructName}}'s actual table name
func TableNameOf{{.StructName}}(shardKey int64) string {
	return TableName{{.StructName}} + strconv.FormatInt(shardKey%TableCount{{.StructName}}, 10)
}
{{end}}
`
