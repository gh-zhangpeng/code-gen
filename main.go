package main

import (
	"code-gen/mysql_gen"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	dsn := "root:Aa40303991.@tcp(127.0.0.1:3306)/box?charset=utf8mb4&parseTime=True&loc=Local"
	db, _ := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	g := mysql_gen.NewGenerator(db, mysql_gen.Config{
		ModelOutputPath: "./files",
	})
	g.GenerateModel("tblMedical")
	//g.GenerateAllTable()
	g.Execute()
}
