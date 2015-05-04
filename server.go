package main

import (
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/learnin/go-multilog"
	"github.com/mattn/go-colorable"
	"github.com/zenazn/goji"
	"github.com/zenazn/goji/graceful"
	"github.com/zenazn/goji/web"
	"github.com/zenazn/goji/web/middleware"

	"github.com/learnin/batch-rest-controller/controllers"
	"github.com/learnin/batch-rest-controller/helpers"
)

const LOG_DIR = "log"
const LOG_FILE = LOG_DIR + "/server.log"

func main() {
	var log *multilog.MultiLogger
	if fi, err := os.Stat(LOG_DIR); os.IsNotExist(err) {
		if err := os.MkdirAll(LOG_DIR, 0755); err != nil {
			panic(err)
		}
	} else {
		if !fi.IsDir() {
			panic("ログディレクトリ " + LOG_DIR + " はディレクトリではありません。")
		}
	}
	logf, err := os.OpenFile(LOG_FILE, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		panic(err)
	}
	stdOutLogrus := logrus.New()
	stdOutLogrus.Out = colorable.NewColorableStdout()
	fileLogrus := logrus.New()
	fileLogrus.Out = logf
	fileLogrus.Formatter = &logrus.TextFormatter{DisableColors: true}
	log = multilog.New(stdOutLogrus, fileLogrus)

	var ds helpers.DataSource
	if err := ds.Connect(); err != nil {
		panic(err)
	}

	jobs := web.New()
	goji.Handle("/jobs/*", jobs)
	jobs.Use(middleware.SubRouter)
	jobsController := controllers.JobsController{DS: &ds, Logger: log}
	jobs.Post("/run", jobsController.Run)
	jobs.Get("/:jobId", jobsController.Show)

	graceful.PostHook(func() {
		if err := ds.Close(); err != nil {
			log.Errorln(err)
		}
		logf.Close()
	})

	goji.Serve()
}
