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
	if d := db.Exec("create table jobs(id integer NOT NULL PRIMARY KEY AUTOINCREMENT, command text NOT NULL, args text, status integer NOT NULL, exit_status integer, created_at TIMESTAMP NOT NULL DEFAULT (DATETIME('now','localtime')), finished_at TIMESTAMP)"); d.Error != nil {
		panic(d.Error)
	}
	if d := db.Exec("create table job_messages(job_id integer NOT NULL, seq integer NOT NULL, type integer NOT NULL, message text, created_at TIMESTAMP NOT NULL DEFAULT (DATETIME('now','localtime')), PRIMARY KEY(job_id, seq))"); d.Error != nil {
		panic(d.Error)
	}
}
