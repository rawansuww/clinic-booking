package models

import (
	"fmt"

	_ "github.com/go-sql-driver/mysql"

	//"github.com/jinzhu/gorm"
	//_ "github.com/jinzhu/gorm/dialects/sqlite"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDatabase() {
	//DO NOT USE CONNECTION NAME IN dsn, but the schema/db name itself!!!
	dsn := "root:asd123asd@tcp(127.0.0.1:3306)/nice?parseTime=true"
	//db, err := sql.Open("mysql", dsn)
	//database, err := gorm.Open("sqlite3", "test.db")
	//dsn := "root:asd123asd@tcp(127.0.0.1:3306)/testDB?charset=utf8mb4&parseTime=True&loc=Local"
	database, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})

	if err != nil {
		fmt.Println(err.Error())
		panic("Failed to connect to database!")
	}

	database.AutoMigrate(&Doctor{}, &Appointment{}, &Patient{}, &Admin{})
	//database.CreateTable(&Doctor{}, &Appointment{})
	//database.Model(&game).Related(&gameImages)
	DB = database
}
