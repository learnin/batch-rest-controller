package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

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

type Job struct {
	Id            int64
	Async         bool
	RequireResult bool
	Command       string
	Args          string
	CreatedAt     time.Time
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
		job := Job{
			Async:         req.Async,
			RequireResult: req.RequireResult,
			Command:       req.Command,
			Args:          req.Args,
		}
		if err := controller.DS.GetDB().Save(&job).Error; err != nil {
			controller.Logger.Errorf("jobs テーブル登録時にエラーが発生しました。error=%v", err)
			sendEroorResponse(w, err, "")
			return
		}
		fmt.Fprintf(w, "{\"id\": %v}", job.Id)
		return
	}
	encoder := json.NewEncoder(w)
	encoder.Encode(response{Error: false, Messages: []string{}})
}
