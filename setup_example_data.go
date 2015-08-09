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
	apiKey := controllers.ApiKey{
		ClientName: "example",
		ApiKey:     "examplekey",
	}
	if d := db.Create(&apiKey); d.Error != nil {
		panic(d.Error)
	}
}
