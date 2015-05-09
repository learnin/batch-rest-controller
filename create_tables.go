package main

import (
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
	if d := db.Exec("create table jobs(id integer NOT NULL PRIMARY KEY AUTOINCREMENT, async integer NOT NULL, require_result integer NOT NULL, command text NOT NULL, args text, created_at TIMESTAMP NOT NULL DEFAULT (DATETIME('now','localtime')))"); d.Error != nil {
		panic(d.Error)
	}
}
