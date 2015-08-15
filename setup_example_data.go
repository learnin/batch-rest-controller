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
	apiKey := models.ApiKey{
		ClientName: "example",
		ApiKey:     "examplekey",
	}
	if d := db.Create(&apiKey); d.Error != nil {
		panic(d.Error)
	}
}
