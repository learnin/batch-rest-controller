package controllers

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"syscall"
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
		out := make(chan string)
		errout := make(chan string)
		jobquit := make(chan bool)

		go func() {
		loop:
			for {
				select {
				case <-jobquit:
					break loop
				case stdout := <-out:
					// TODO テーブルに格納する
					fmt.Println(stdout)
				case stderr := <-errout:
					// TODO テーブルに格納する
					fmt.Println(stderr)
				}
			}
		}()
		go func() {
			cmd := exec.Command("ps", "-ef")
			stdout, err := cmd.StdoutPipe()
			if err != nil {
				// TODO テーブルに登録する
				fmt.Println(err)
				jobquit <- true
			}
			stderr, err := cmd.StderrPipe()
			if err != nil {
				// TODO テーブルに登録する
				fmt.Println(err)
				jobquit <- true
			}

			if err := cmd.Start(); err != nil {
				// TODO テーブルに登録する
				fmt.Println(err)
				jobquit <- true
			}

			go func() {
				scanner := bufio.NewScanner(stdout)
				for scanner.Scan() {
					out <- scanner.Text()
				}
			}()
			go func() {
				scanner := bufio.NewScanner(stderr)
				for scanner.Scan() {
					errout <- scanner.Text()
				}
			}()
			if err := cmd.Wait(); err != nil {
				if err2, ok := err.(*exec.ExitError); ok {
					if s, ok := err2.Sys().(syscall.WaitStatus); ok {
						fmt.Println(err)
						// TODO テーブルに登録する
						fmt.Println(s.ExitStatus())
					} else {
						// Unix や Winodws とは異なり、 exec.ExitError.Sys() が syscall.WaitStatus ではないOSの場合
						fmt.Println(err)
					}
				} else {
					// may be returned for I/O problems.
					fmt.Println(err)
				}
			} else {
				// TODO exitStatus = 0. テーブルに登録する
				fmt.Println(0)
			}
			jobquit <- true
		}()
		fmt.Fprintf(w, "{\"id\": %v}", job.Id)
		return
	}
	encoder := json.NewEncoder(w)
	encoder.Encode(response{Error: false, Messages: []string{}})
}
