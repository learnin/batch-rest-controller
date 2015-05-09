package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/learnin/go-multilog"
	"github.com/zenazn/goji/web"

	"github.com/learnin/batch-rest-controller/helpers"
)

type JobsController struct {
	DS     *helpers.DataSource
	Logger *multilog.MultiLogger
}

type Request struct {
	Async         bool
	RequireResult bool
	Command       string
	Args          string
}

func (controller *JobsController) Show(c web.C, w http.ResponseWriter, r *http.Request) {
	// jobId := c.URLParams["jobId"]
	encoder := json.NewEncoder(w)
	encoder.Encode(response{Error: false, Messages: []string{}})
}

func (controller *JobsController) Run(c web.C, w http.ResponseWriter, r *http.Request) {
	req := Request{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		controller.Logger.Error(err)
		http.Error(w, "", http.StatusBadRequest)
		return
	}
	if req.Async && req.RequireResult {
		db := controller.DS.GetDB()
		if d := db.Exec("insert into jobs(async, require_result, command, args) VALUES(?, ?, ?, ?)", req.Async, req.RequireResult, req.Command, req.Args); d.Error != nil {
			controller.Logger.Errorf("jobs テーブル登録時にエラーが発生しました。error=%v", d.Error)
			sendEroorResponse(w, d.Error, "")
			return
		}
	}
	encoder := json.NewEncoder(w)
	encoder.Encode(response{Error: false, Messages: []string{}})
}
