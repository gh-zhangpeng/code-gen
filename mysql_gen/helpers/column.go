package helpers

import (
	"bytes"
	"fmt"
	"gorm.io/gorm"
	"reflect"
	"strings"
)

// Index table index info
type Index struct {
	gorm.Index
	Priority int32 `gorm:"column:SEQ_IN_INDEX"`
}

// Column table column's info
type Column struct {
	gormColumnType gorm.ColumnType
	Table          string       //tblUser
	Name           string       //user_id
	Kind           reflect.Kind //reflect.String
	Type           string       //varchar
	DetailType     string       //varchar(255)
	DefaultValue   string       //0
	Comment        string       //用户ID
	Nullable       bool
	Indexes        []*Index
}

func (c *Column) BuildGormTag() string {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("column:%s;type:%s", c.Name, c.DetailType))
	isPriKey, ok := c.gormColumnType.PrimaryKey()
	isValidPriKey := ok && isPriKey
	if isValidPriKey {
		buf.WriteString(";primaryKey")
		if at, ok := c.gormColumnType.AutoIncrement(); ok {
			buf.WriteString(fmt.Sprintf(";autoIncrement:%t", at))
		}
	} else if n, ok := c.gormColumnType.Nullable(); ok && !n {
		buf.WriteString(";not null")
	}
	for _, idx := range c.Indexes {
		if idx == nil {
			continue
		}
		if pk, _ := idx.PrimaryKey(); pk { //ignore PrimaryKey
			continue
		}
		if uniq, _ := idx.Unique(); uniq {
			buf.WriteString(fmt.Sprintf(";uniqueIndex:%s,priority:%d", idx.Name(), idx.Priority))
		} else {
			buf.WriteString(fmt.Sprintf(";index:%s,priority:%d", idx.Name(), idx.Priority))
		}
	}
	if !isValidPriKey && len(strings.TrimSpace(c.DefaultValue)) > 0 {
		buf.WriteString(fmt.Sprintf(`;default:'%s'`, c.DefaultValue))
	}
	return buf.String()
}
