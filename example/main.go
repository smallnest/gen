package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"github.com/smallnest/gen/example/api"
)

// gin
func GinServer() {
	router := gin.Default()
	api.ConfigGinRouter(router)
	router.Run(":8080")
}

func main() {
	db, err := gorm.Open("mysql", "root@/employees?charset=utf8&parseTime=True")
	if err != nil {
		log.Fatalf("Got error when connect database, the error is '%v'", err)
	}

	db.LogMode(true)
	api.DB = db

	// db.AutoMigrate(&model.Department{}, &model.DeptEmp{}, &model.DeptManager{}, &model.Employee{},
	// 	&model.Salary{}, &model.Title{})

	go GinServer()

	r := api.ConfigRouter()
	log.Fatal(http.ListenAndServe(":8899", r))
}
