package mysql_gen

import (
	"fmt"
	"gorm.io/gorm"
	"regexp"
	"strings"
)

//	type Config struct {
//		db *gorm.DB // db connection
//
//		OutPath      string // query code path
//		OutFile      string // query code file name, default: gen.go
//		ModelPkgPath string // generated model code's package name
//		WithUnitTest bool   // generate unit test for query code
//
//		// generate model global configuration
//		FieldNullable     bool // generate pointer when field is nullable
//		FieldCoverable    bool // generate pointer when field has default value, to fix problem zero value cannot be assign: https://gorm.io/docs/create.html#Default-Values
//		FieldSignable     bool // detect integer field's unsigned type, adjust generated data type
//		FieldWithIndexTag bool // generate with gorm index tag
//		FieldWithTypeTag  bool // generate with gorm column type tag
//
//		Mode GenerateMode // generate mode
//
//		queryPkgName   string // generated query code's package name
//		modelPkgPath   string // model pkg path in target project
//		dbNameOpts     []model.SchemaNameOpt
//		importPkgPaths []string
//
//		// name strategy for syncing table from db
//		tableNameNS func(tableName string) (targetTableName string)
//		modelNameNS func(tableName string) (modelName string)
//		fileNameNS  func(tableName string) (fileName string)
//
//		dataTypeMap    map[string]func(detailType string) (dataType string)
//		fieldJSONTagNS func(columnName string) (tagContent string)
//		fieldNewTagNS  func(columnName string) (tagContent string)
//	}
type Config struct {
	//需要生成结构的数据库链接
	Database *gorm.DB
	//生成文件的位置
	OutFilePath string
}
type Generator struct {
	config Config
	tables []string
}

func NewGenerator(config Config) *Generator {
	return &Generator{config: config}
}

// GenerateTable 生成指定表的结构体
func (g *Generator) GenerateTable(table string) {
	if !g.config.Database.Migrator().HasTable(table) {
		panic(fmt.Errorf("%s does not exist", table))
	}
	g.tables = []string{table}
}

// GenerateAllTable 生成所有表的结构体
func (g *Generator) GenerateAllTable() {
	g.tables = getTables(g.config.Database)
}

// Execute 执行
func (g *Generator) Execute() {
	fmt.Printf("The following tables struct will be generated: %+v\n", g.tables)
	fmt.Println("Start generating code")
	for _, tableName := range g.tables {
		formatTable2Struct(tableName)
	}
	fmt.Println("Generate code done")
}

var tableNameRegex = regexp.MustCompile("^([A-Za-z][-_]*)+[0-9]*$")
var letterRegex = regexp.MustCompile("[A-Za-z]+")

func formatTable2Struct(table string) error {
	if len(table) == 0 {
		return fmt.Errorf("tableName is empty")
	}
	if !tableNameRegex.MatchString(table) {
		return fmt.Errorf("tableName contains invalid characters")
	}
	formatTableName2StructName := func(s string) string {
		matchedStrings := letterRegex.FindAllString(s, -1)
		if len(matchedStrings) == 0 {
			return ""
		}
		var structName string
		for _, matchedString := range matchedStrings {
			structName = fmt.Sprintf("%s%s", structName, strings.ToUpper(matchedString[:1])+matchedString[1:])
		}
		return structName
	}
	structName := formatTableName2StructName(table)
	fmt.Println(structName)
	return nil
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
