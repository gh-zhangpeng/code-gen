package helpers

import (
	"gorm.io/gorm"
	"reflect"
	"strings"
)

// CheckTableAvailability 检查表是否存在
func CheckTableAvailability(db *gorm.DB, table string) bool {
	return db.Migrator().HasTable(table)
}

type Table struct {
	Columns      []Column
	tableIndexes []gorm.Index
	tableColumns []gorm.ColumnType
}

func groupIndexesByColumn(indexList []gorm.Index) map[string][]*Index {
	columnIndexMap := make(map[string][]*Index, len(indexList))
	if len(indexList) == 0 {
		return columnIndexMap
	}

	for _, idx := range indexList {
		if idx == nil {
			continue
		}
		for i, col := range idx.Columns() {
			columnIndexMap[col] = append(columnIndexMap[col], &Index{
				Index:    idx,
				Priority: int32(i + 1),
			})
		}
	}
	return columnIndexMap
}

// GetTableInfo get table info
func GetTableInfo(db *gorm.DB, table string) (*Table, error) {
	tableColumns, err := getTableColumns(db, table)
	if err != nil {
		return nil, err
	}
	columns := make([]Column, 0, len(tableColumns))
	for _, tableColumn := range tableColumns {
		detailType, _ := tableColumn.ColumnType()
		defaultValue, _ := tableColumn.DefaultValue()
		nullable, _ := tableColumn.Nullable()
		comment, _ := tableColumn.Comment()
		column := Column{
			gormColumnType: tableColumn,
			Table:          table,
			Name:           tableColumn.Name(),
			Kind:           tableColumn.ScanType().Kind(),
			Type:           tableColumn.DatabaseTypeName(),
			DetailType:     detailType,
			DefaultValue:   defaultValue,
			Comment:        comment,
			Nullable:       nullable,
			Indexes:        nil,
		}
		columns = append(columns, column)
	}
	tableIndexes, err := getTableIndexes(db, table)
	if err != nil {
		return nil, err
	}
	column2Indexes := groupIndexesByColumn(tableIndexes)
	for i, column := range columns {
		v, ok := column2Indexes[column.Name]
		if !ok {
			continue
		}
		columns[i].Indexes = v
	}
	return &Table{
		Columns:      columns,
		tableColumns: tableColumns,
		tableIndexes: tableIndexes,
	}, nil
}

// needDefaultTag check if default tag needed
func needDefaultTag(kind reflect.Kind, defaultTagValue string) bool {
	if defaultTagValue == "" {
		return false
	}
	switch kind {
	case reflect.Bool:
		return defaultTagValue != "false"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64:
		return defaultTagValue != "0"
	case reflect.String:
		return defaultTagValue != ""
	case reflect.Struct:
		return strings.Trim(defaultTagValue, "'0:- ") != ""
	}
	return false
}

func getTableColumns(db *gorm.DB, tableName string) ([]gorm.ColumnType, error) {
	return db.Migrator().ColumnTypes(tableName)
}

func getTableIndexes(db *gorm.DB, tableName string) (indexes []gorm.Index, err error) {
	return db.Migrator().GetIndexes(tableName)
}
