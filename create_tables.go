package main

import (
	"github.com/learnin/batch-rest-controller/helpers"
	"github.com/learnin/batch-rest-controller/models"
)

func main() {
	var ds helpers.DataSource
	if err := ds.Connect(); err != nil {
		panic(err)
	}
	defer ds.Close()
	db := ds.GetDB()
	db.LogMode(true)
	if d := db.AutoMigrate(&models.Job{}); d.Error != nil {
		panic(d.Error)
	}
	if d := db.AutoMigrate(&models.JobMessage{}); d.Error != nil {
		panic(d.Error)
	}
	if d := db.AutoMigrate(&models.ApiKey{}); d.Error != nil {
		panic(d.Error)
	}

}
