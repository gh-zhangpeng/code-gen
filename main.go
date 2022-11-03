package main

import (
	"code-gen/mysql_gen"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	dsn := "root:Aa40303991.@tcp(127.0.0.1:3306)/box?charset=utf8mb4&parseTime=True&loc=Local"
	db, _ := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	g := mysql_gen.NewGenerator(mysql_gen.Config{
		Database:    db,
		OutFilePath: "./files",
	})
	g.GenerateTable("tblUser")
	//g.GenerateAllTable()
	g.Execute()
}
