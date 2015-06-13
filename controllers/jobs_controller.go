package controllers

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
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

type JobMessage struct {
	JobId     int64
	Seq       int64
	Type      int
	Message   string
	CreatedAt time.Time
}

func (controller *JobsController) Show(c web.C, w http.ResponseWriter, r *http.Request) {
	jobId := c.URLParams["jobId"]
	job := Job{}
	// FIXME SQLiteのマルチスレッドサポートは接続単位なので、都度Open,Closeする
	if d := controller.DS.GetDB().First(&job, jobId); d.Error != nil {
		if d.RecordNotFound() {
			http.Error(w, "", http.StatusNotFound)
			return
		}
		controller.Logger.Errorf("jobs テーブル取得時にエラーが発生しました。error=%v", d.Error)
		sendEroorResponse(w, d.Error, "")
		return
	}
	msgs := []JobMessage{}
	resMsgs := []string{}
	if d := controller.DS.GetDB().Order("seq").Find(&msgs, "job_id = ?", jobId).Pluck("message", &resMsgs); d.Error != nil {
		if d.RecordNotFound() {
			encoder := json.NewEncoder(w)
			encoder.Encode(response{Error: false, Messages: []string{}})
			return
		}
		controller.Logger.Errorf("jobMessages テーブル取得時にエラーが発生しました。error=%v", d.Error)
		sendEroorResponse(w, d.Error, "")
		return
	}
	encoder := json.NewEncoder(w)
	encoder.Encode(response{Error: false, Messages: resMsgs})
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
		cmd := exec.Command("ps", "-ef")
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			controller.Logger.Errorf("標準出力パイプ取得時にエラーが発生しました。error=%v", err)
			sendEroorResponse(w, err, "")
			return
		}
		stderr, err := cmd.StderrPipe()
		if err != nil {
			controller.Logger.Errorf("標準エラー出力パイプ取得時にエラーが発生しました。error=%v", err)
			sendEroorResponse(w, err, "")
			return
		}

		go func(stdout *io.ReadCloser, stderr *io.ReadCloser) {
			if err := cmd.Start(); err != nil {
				jobMsg := JobMessage{
					JobId:   job.Id,
					Seq:     1,
					Type:    2,
					Message: err.Error(),
				}
				if err := controller.DS.GetDB().Save(&jobMsg).Error; err != nil {
					controller.Logger.Errorf("job_messages テーブル登録時にエラーが発生しました。error=%v", err)
					return
				}
				return
			}

			out := make(chan string)
			errout := make(chan string)
			jobquit := make(chan bool)

			go func() {
				nowSeq := int64(0)
			loop:
				for {
					select {
					case <-jobquit:
						break loop
					case stdout := <-out:
						nowSeq++
						jobMsg := JobMessage{
							JobId:   job.Id,
							Seq:     nowSeq,
							Type:    1,
							Message: stdout,
						}
						if err := controller.DS.GetDB().Save(&jobMsg).Error; err != nil {
							controller.Logger.Errorf("job_messages テーブル登録時にエラーが発生しました。error=%v", err)
							return
						}
					case stderr := <-errout:
						nowSeq++
						jobMsg := JobMessage{
							JobId:   job.Id,
							Seq:     nowSeq,
							Type:    2,
							Message: stderr,
						}
						if err := controller.DS.GetDB().Save(&jobMsg).Error; err != nil {
							controller.Logger.Errorf("job_messages テーブル登録時にエラーが発生しました。error=%v", err)
							return
						}
					}
				}
			}()
			go func() {
				scanner := bufio.NewScanner(*stdout)
				for scanner.Scan() {
					out <- scanner.Text()
				}
			}()
			go func() {
				scanner := bufio.NewScanner(*stderr)
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
		}(&stdout, &stderr)

		fmt.Fprintf(w, "{\"id\": %v}", job.Id)
		return
	} else if !req.Async {
		cmd := exec.Command("ps", "-ef")
		var stdout bytes.Buffer
		var stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			if err2, ok := err.(*exec.ExitError); ok {
				if s, ok := err2.Sys().(syscall.WaitStatus); ok {
					sendEroorResponse(w, err, fmt.Sprintf("バッチが正常終了しませんでした。exitStatus=%v stdout=%v stderr=%v", s.ExitStatus(), stdout.String(), stderr.String()))
					return
				} else {
					// Unix や Winodws とは異なり、 exec.ExitError.Sys() が syscall.WaitStatus ではないOSの場合
					sendEroorResponse(w, err, fmt.Sprintf("バッチが正常終了しませんでした。stdout=%v stderr=%v", stdout.String(), stderr.String()))
					return
				}
			} else {
				// may be returned for I/O problems.
				sendEroorResponse(w, err, "バッチ実行時にエラーが発生しました。")
				return
			}
		}
		fmt.Fprintf(w, "{\"exitStatus\": 0, \"stdout\": \"%v\", \"stderr\": \"%v\"}", stdout.String(), stderr.String())
		return
	}
	encoder := json.NewEncoder(w)
	encoder.Encode(response{Error: false, Messages: []string{}})
}
