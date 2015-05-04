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

type ReqForm struct {
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
	form := ReqForm{}
	if err := json.NewDecoder(r.Body).Decode(&form); err != nil {
		http.Error(w, "", http.StatusBadRequest)
		return
	}
	encoder := json.NewEncoder(w)
	encoder.Encode(response{Error: false, Messages: []string{}})
}
