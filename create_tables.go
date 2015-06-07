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
	if d := db.Exec("create table job_messages(job_id integer NOT NULL, seq integer NOT NULL, type integer NOT NULL, message text, created_at TIMESTAMP NOT NULL DEFAULT (DATETIME('now','localtime')), PRIMARY KEY(job_id, seq))"); d.Error != nil {
		panic(d.Error)
	}
}
