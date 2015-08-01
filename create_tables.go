package main

import (
	"github.com/learnin/batch-rest-controller/controllers"
	"github.com/learnin/batch-rest-controller/helpers"
)

func main() {
	var ds helpers.DataSource
	if err := ds.Connect(); err != nil {
		panic(err)
	}
	defer ds.Close()
	db := ds.GetDB()
	db.LogMode(true)
	if d := db.AutoMigrate(&controllers.Job{}); d.Error != nil {
		panic(d.Error)
	}
	if d := db.AutoMigrate(&controllers.JobMessage{}); d.Error != nil {
		panic(d.Error)
	}
	if d := db.AutoMigrate(&controllers.ApiKey{}); d.Error != nil {
		panic(d.Error)
	}

}
