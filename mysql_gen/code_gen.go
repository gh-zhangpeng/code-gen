package mysql_gen

import (
	"bytes"
	"code-gen/mysql_gen/helpers"
	"code-gen/mysql_gen/templates"
	"code-gen/pool"
	"fmt"
	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/imports"
	"gorm.io/gorm"
	"html/template"
	"io"
	"os"
	"reflect"
	"regexp"
	"runtime"
	"strconv"
	"strings"
)

const defaultModelOutputPath = "./dal/model/"

type Generator struct {
	db        *gorm.DB
	config    Config
	tables    []string
	structMap map[string]*StructMeta //gen model data
}

// NewGenerator init generator
func NewGenerator(db *gorm.DB, config Config) *Generator {
	if len(config.ModelOutputPath) == 0 {
		config.ModelOutputPath = defaultModelOutputPath
	}
	return &Generator{db: db, config: config, structMap: map[string]*StructMeta{}}
}

type Field struct {
	Name             string
	Type             string
	Comment          string
	MultilineComment bool
	GORMTag          string
	JSONTag          string
	NewTag           string
}

type StructMeta struct {
	FileName   string // generated file name
	StructName string // origin/model struct name
	//TableName  string // table name in db server
	Fields []Field
	Tables []string
}

// GenerateModel 生成指定表的结构体
func (g *Generator) GenerateModel(table string) (*StructMeta, error) {
	return g.GenerateModelWithOption(table, strings.ToUpper(table[:1])+table[1:], StructConfig{})
}
func (g *Generator) GenerateModelWithOption(table string, structName string, config StructConfig) (*StructMeta, error) {
	if !helpers.CheckTableAvailability(g.db, table) {
		return nil, fmt.Errorf("%s does not exist\n", table)
	}
	tableInfo, err := helpers.GetTableInfo(g.db, table)
	if err != nil {
		return nil, err
	}
	_, structName, fileName, err := getNames(table)
	if err != nil {
		return nil, err
	}
	if config.structNameNS != nil {
		structName = config.structNameNS(table)
	}
	if config.fileNameNS != nil {
		fileName = config.fileNameNS(table)
	}
	if len(structName) == 0 {
		return nil, fmt.Errorf("newTableName is empty")
	}
	if !firstUpperRegex.MatchString(structName) {
		return nil, fmt.Errorf("struct name %s is invalid", structName)
	}
	if len(fileName) == 0 {
		return nil, fmt.Errorf("file is empty")
	}

	if _, ok := g.structMap[structName]; ok {
		g.structMap[structName].Tables = append(g.structMap[structName].Tables, table)
	} else {
		g.structMap[structName] = &StructMeta{
			FileName:   fileName,
			StructName: structName,
			Fields:     getFields(tableInfo, config.FiledConfig),
			Tables:     []string{},
		}
	}
	return g.structMap[structName], nil
}

const defaultFieldType = "string"

var (
	defaultColumnType2GoType dataTypeMap = map[string]dataTypeMapping{
		"numeric":    func(string) string { return "int32" },
		"integer":    func(string) string { return "int32" },
		"int":        func(string) string { return "int32" },
		"smallint":   func(string) string { return "int32" },
		"mediumint":  func(string) string { return "int32" },
		"bigint":     func(string) string { return "int64" },
		"float":      func(string) string { return "float32" },
		"real":       func(string) string { return "float64" },
		"double":     func(string) string { return "float64" },
		"decimal":    func(string) string { return "float64" },
		"char":       func(string) string { return "string" },
		"varchar":    func(string) string { return "string" },
		"tinytext":   func(string) string { return "string" },
		"mediumtext": func(string) string { return "string" },
		"longtext":   func(string) string { return "string" },
		"binary":     func(string) string { return "[]byte" },
		"varbinary":  func(string) string { return "[]byte" },
		"tinyblob":   func(string) string { return "[]byte" },
		"blob":       func(string) string { return "[]byte" },
		"mediumblob": func(string) string { return "[]byte" },
		"longblob":   func(string) string { return "[]byte" },
		"text":       func(string) string { return "string" },
		"json":       func(string) string { return "string" },
		"enum":       func(string) string { return "string" },
		"time":       func(string) string { return "time.Time" },
		"date":       func(string) string { return "time.Time" },
		"datetime":   func(string) string { return "time.Time" },
		"timestamp":  func(string) string { return "time.Time" },
		"year":       func(string) string { return "int32" },
		"bit":        func(string) string { return "[]uint8" },
		"boolean":    func(string) string { return "bool" },
		"tinyint": func(detailType string) string {
			if strings.HasPrefix(strings.TrimSpace(detailType), "tinyint(1)") {
				return "bool"
			}
			return "int32"
		},
	}
)

type dataTypeMapping func(detailType string) (finalType string)

type dataTypeMap map[string]dataTypeMapping

func (m dataTypeMap) Get(dataType, detailType string) string {
	if convert, ok := m[strings.ToLower(dataType)]; ok {
		return convert(detailType)
	}
	return defaultFieldType
}

// needDefaultTag check if default tag needed
func needDefaultTag(columnDefaultType reflect.Kind, columnDefaultValue string) bool {
	switch columnDefaultType {
	case reflect.Bool:
		return columnDefaultValue != "false"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64:
		return columnDefaultValue != "0"
	case reflect.String:
		return columnDefaultValue != ""
	case reflect.Struct:
		return strings.Trim(columnDefaultValue, "'0:- ") != ""
	default:
		return false
	}
}

func formatColumn2Field(column helpers.Column, config FieldConfig) Field {
	fieldType := defaultFieldType
	//1.默认的数据库类型与 Go 类型的对应关系
	if convert, ok := defaultColumnType2GoType[strings.ToLower(column.Type)]; ok {
		fieldType = convert(column.Type)
	}
	//2.如果是需要符号，则通过详细类型判断是否需要添加 u
	if config.FieldSigned &&
		strings.Contains(column.DetailType, "unsigned") &&
		strings.HasPrefix(fieldType, "int") {
		fieldType = "u" + fieldType
	}
	switch {
	case column.Name == "deleted_at" && fieldType == "time.Time":
		fieldType = "gorm.DeletedAt"
	case config.FieldCoverable && needDefaultTag(column.Kind, column.DefaultValue):
		fieldType = "*" + fieldType
	case config.FieldNullable && column.Nullable:
		fieldType = "*" + fieldType
	}
	//3.用户指定类型映射关系
	if mapping, ok := config.dataTypeMap[column.Type]; ok {
		//获取字段的真实类型（如：varchar(64)），然后传递给用户指定的映射关系进行处理
		if len(column.DetailType) > 0 {
			fieldType = mapping(column.DetailType)
		}
	}
	field := Field{
		Name:             column.Name,
		Type:             fieldType,
		Comment:          column.Comment,
		MultilineComment: strings.Contains(column.Comment, "\n"),
		GORMTag:          column.BuildGormTag(),
	}
	if config.fieldJSONTagNS != nil {
		field.JSONTag = config.fieldJSONTagNS(column.Name)
	}
	if config.fieldNewTagNS != nil {
		field.NewTag = config.fieldNewTagNS(column.Name)
	}
	if !config.FieldWithTypeTag { // remove type tag if FieldWithTypeTag == false
		field.GORMTag = strings.ReplaceAll(field.GORMTag, ";type:"+column.DetailType, "")
	}
	return field
}

func getFields(table *helpers.Table, config FieldConfig) []Field {
	fields := make([]Field, 0, len(table.Columns))
	for _, col := range table.Columns {
		fields = append(fields, formatColumn2Field(col, config))
	}
	return fields
}

// GenerateAllTable 生成所有表的结构体
func (g *Generator) GenerateAllTable() {
	g.tables = getTables(g.db)
}

// Execute 执行
func (g *Generator) Execute() {
	fmt.Printf("The following tables struct will be generated: %+v\n", g.tables)
	fmt.Println("Start generating code")
	g.generateModelFile()
	fmt.Println("Generate code done")
}

// generateModelFile generate model structures and save to file
func (g *Generator) generateModelFile() error {
	if len(g.structMap) == 0 {
		return nil
	}

	if err := os.MkdirAll(g.config.ModelOutputPath, os.ModePerm); err != nil {
		return fmt.Errorf("create model pkg path(%s) fail: %s", g.config.ModelOutputPath, err)
	}
	errChan := make(chan error)
	var concurrent = runtime.NumCPU()
	p := pool.NewPool(concurrent)
	for _, data := range g.structMap {
		if data == nil {
			continue
		}
		p.Wait()
		go func(data *StructMeta) {
			defer p.Done()

			var buf bytes.Buffer
			err := render(templates.Model, &buf, data)
			if err != nil {
				errChan <- err
				return
			}

			modelFile := g.config.ModelOutputPath + data.FileName + ".gen.go"
			err = output(modelFile, buf.Bytes())
			if err != nil {
				errChan <- err
				return
			}

			//g.info(fmt.Sprintf("generate model file(table <%s> -> {%s.%s}): %s", data.TableName, data.StructInfo.Package, data.StructInfo.Type, modelFile))
		}(data)
	}
	select {
	case err := <-errChan:
		return err
	case <-p.AsyncWaitAll():
		g.fillModelPkgPath(g.config.ModelOutputPath)
	}
	return nil
}

func (g *Generator) fillModelPkgPath(filePath string) {
	pkgs, err := packages.Load(&packages.Config{
		Mode: packages.NeedName,
		Dir:  filePath,
	})
	if err != nil {
		return
	}
	if len(pkgs) == 0 {
		return
	}
	fmt.Println(pkgs[0].PkgPath)
}

func output(fileName string, content []byte) error {
	result, err := imports.Process(fileName, content, nil)
	if err != nil {
		lines := strings.Split(string(content), "\n")
		errLine, _ := strconv.Atoi(strings.Split(err.Error(), ":")[1])
		startLine, endLine := errLine-5, errLine+5
		fmt.Println("Format fail:", errLine, err)
		if startLine < 0 {
			startLine = 0
		}
		if endLine > len(lines)-1 {
			endLine = len(lines) - 1
		}
		for i := startLine; i <= endLine; i++ {
			fmt.Println(i, lines[i])
		}
		return fmt.Errorf("cannot format file: %w", err)
	}
	return os.WriteFile(fileName, result, 0640)
}

func render(tmpl string, wr io.Writer, data interface{}) error {
	t, err := template.New(tmpl).Parse(tmpl)
	if err != nil {
		return err
	}
	err = t.Execute(wr, data)
	if err != nil {
		return err
	}
	return nil
}

var tableNameRegex = regexp.MustCompile("^([A-Za-z][-_]*)+[0-9]*$")
var letterRegex = regexp.MustCompile("[A-Za-z]+")
var firstUpperRegex = regexp.MustCompile("[A-Z]+[a-z]*")

func formatTableName2NewTableName(tableName string) string {
	matchedStrings := letterRegex.FindAllString(tableName, -1)
	if len(matchedStrings) == 0 {
		return ""
	}
	var newTableName string
	for _, matchedString := range matchedStrings {
		newTableName = fmt.Sprintf("%s%s", newTableName, matchedString)
	}
	return newTableName
}

func formatTableName2StructName(tableName string) string {
	matchedStrings := letterRegex.FindAllString(tableName, -1)
	if len(matchedStrings) == 0 {
		return ""
	}
	var structName string
	for _, matchedString := range matchedStrings {
		structName = fmt.Sprintf("%s%s", structName, strings.ToUpper(matchedString[:1])+matchedString[1:])
	}
	return structName
}

func formatStructName2FileName(structName string) string {
	matchedStrings := firstUpperRegex.FindAllString(structName, -1)
	if len(matchedStrings) == 0 {
		return ""
	}
	var fileName string
	for i, matchedString := range matchedStrings {
		fileName = fileName + strings.ToLower(matchedString)
		if i != len(matchedStrings)-1 {
			fileName = fileName + "_"
		}
	}
	return fileName
}

func getNames(table string) (newTableName, structName string, fileName string, err error) {
	if len(table) == 0 {
		return newTableName, structName, fileName, fmt.Errorf("tableName is empty")
	}
	if !tableNameRegex.MatchString(table) {
		return newTableName, structName, fileName, fmt.Errorf("tableName contains invalid characters")
	}
	newTableName = formatTableName2NewTableName(table)
	structName = formatTableName2StructName(table)
	fileName = formatStructName2FileName(structName)
	return newTableName, structName, fileName, nil
}

// getTables 获取所有表
func getTables(database *gorm.DB) []string {
	tables, err := database.Migrator().GetTables()
	if err != nil {
		panic(fmt.Errorf("get all tables fail: %w", err))
	}
	fmt.Printf("find %d table from database, tables: %+v\n", len(tables), tables)
	return tables
}
