package mysql_gen

type Config struct {
	ModelOutputPath string
}

type FieldConfig struct {
	// generate model global configuration
	FieldNullable     bool // generate pointer when field is nullable
	FieldCoverable    bool // generate pointer when field has default value, to fix problem zero value cannot be assign: https://gorm.io/docs/create.html#Default-Values
	FieldSigned       bool // detect integer field's unsigned type, adjust generated data type
	FieldWithIndexTag bool // generate with gorm index tag
	FieldWithTypeTag  bool // generate with gorm column type tag

	dataTypeMap    map[string]func(detailType string) (dataType string)
	fieldJSONTagNS func(columnName string) (tag string)
	fieldNewTagNS  func(columnName string) (tag string)
}

type StructConfig struct {
	FiledConfig FieldConfig
	// name strategy for syncing table from db
	//tableNameNS  func(table string) (newTableName string)
	structNameNS func(table string) (structName string)
	fileNameNS   func(table string) (fileName string)
}

// WithTableNameStrategy specify table name naming strategy, only work when syncing table from db
//func (config *StructConfig) WithTableNameStrategy(ns func(table string) (newTableName string)) {
//	config.tableNameNS = ns
//}

// WithStructNameStrategy specify model struct name naming strategy, only work when syncing table from db
func (config *StructConfig) WithStructNameStrategy(ns func(table string) (modelName string)) {
	config.structNameNS = ns
}

// WithFileNameStrategy specify file name naming strategy, only work when syncing table from db
func (config *StructConfig) WithFileNameStrategy(ns func(table string) (fileName string)) {
	config.fileNameNS = ns
}

// WithDataTypeMap specify data type mapping relationship, only work when syncing table from db
func (config *StructConfig) WithDataTypeMap(newMap map[string]func(detailType string) (dataType string)) {
	config.FiledConfig.dataTypeMap = newMap
}

// WithJSONTagNameStrategy specify json tag naming strategy
func (config *StructConfig) WithJSONTagNameStrategy(ns func(columnName string) (tagContent string)) {
	config.FiledConfig.fieldJSONTagNS = ns
}

// WithNewTagNameStrategy specify new tag naming strategy
func (config *StructConfig) WithNewTagNameStrategy(ns func(columnName string) (tagContent string)) {
	config.FiledConfig.fieldNewTagNS = ns
}
