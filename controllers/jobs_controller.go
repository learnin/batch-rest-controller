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
	"github.com/learnin/batch-rest-controller/models"
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

type showResponse struct {
	response
	Status string
}

func (controller *JobsController) Show(c web.C, w http.ResponseWriter, r *http.Request) {
	jobId := c.URLParams["jobId"]
	job := models.Job{}
	if d := controller.DS.GetDB().First(&job, jobId); d.Error != nil {
		if d.RecordNotFound() {
			http.Error(w, "", http.StatusNotFound)
			return
		}
		controller.Logger.Errorf("jobs テーブル取得時にエラーが発生しました。error=%v", d.Error)
		SendEroorResponse(w, d.Error, "")
		return
	}
	resMsgs := []string{}
	if d := controller.DS.GetDB().Model(&models.JobMessage{}).Where("job_id = ?", jobId).Order("seq").Pluck("message", &resMsgs); d.Error != nil {
		if d.RecordNotFound() {
			encoder := json.NewEncoder(w)
			encoder.Encode(response{Error: false, Messages: []string{}})
			return
		}
		controller.Logger.Errorf("jobMessages テーブル取得時にエラーが発生しました。error=%v", d.Error)
		SendEroorResponse(w, d.Error, "")
		return
	}
	encoder := json.NewEncoder(w)
	encoder.Encode(
		showResponse{
			response: response{
				Error:    false,
				Messages: resMsgs,
			},
			Status: job.Status.String(),
		})
}

func (controller *JobsController) Run(c web.C, w http.ResponseWriter, r *http.Request) {
	req := Request{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		controller.Logger.Error(err)
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	if req.Async && req.RequireResult {
		job := models.Job{
			Command: req.Command,
			Args:    req.Args,
			Status:  models.WaitingToRun,
		}
		if req.RequireResult {
			if err := controller.DS.DoInTransaction(func(th *helpers.TxHolder) error {
				return th.GetTx().Save(&job).Error
			}); err != nil {
				controller.Logger.Errorf("jobs テーブル登録時にエラーが発生しました。error=%v", err)
				SendEroorResponse(w, err, "")
				return
			}
		}
		cmd := exec.Command(job.Command, job.Args)
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			controller.Logger.Errorf("標準出力パイプ取得時にエラーが発生しました。error=%v", err)
			SendEroorResponse(w, err, "")
			return
		}
		stderr, err := cmd.StderrPipe()
		if err != nil {
			controller.Logger.Errorf("標準エラー出力パイプ取得時にエラーが発生しました。error=%v", err)
			SendEroorResponse(w, err, "")
			return
		}
		go func(stdout *io.ReadCloser, stderr *io.ReadCloser) {
			out := make(chan string)
			errout := make(chan string)
			jobquit := make(chan bool)

			stdoutfin := make(chan bool)
			stderrfin := make(chan bool)

			go func() {
				scanner := bufio.NewScanner(*stdout)
				for scanner.Scan() {
					out <- scanner.Text()
				}
				if err := scanner.Err(); err != nil {
					errout <- fmt.Sprintf("標準出力の読み取り時にエラーが発生しました。error=%v", err)
				}
				stdoutfin <- true
			}()
			go func() {
				scanner := bufio.NewScanner(*stderr)
				for scanner.Scan() {
					errout <- scanner.Text()
				}
				if err := scanner.Err(); err != nil {
					errout <- fmt.Sprintf("標準エラー出力の読み取り時にエラーが発生しました。error=%v", err)
				}
				stderrfin <- true
			}()
			if err := cmd.Start(); err != nil && req.RequireResult {
				jobMsg := models.JobMessage{
					JobId:   job.Id,
					Seq:     1,
					Type:    models.Error,
					Message: err.Error(),
				}
				controller.insertJobMessage(&jobMsg)
				return
			}
			job.Status = models.Running
			if err := controller.DS.DoInTransaction(func(th *helpers.TxHolder) error {
				return th.GetTx().Save(&job).Error
			}); err != nil {
				controller.Logger.Errorf("jobs テーブル更新時にエラーが発生しました。error=%v", err)
				return
			}
			go func() {
				nowSeq := int64(0)
			loop:
				for {
					select {
					case <-jobquit:
						break loop
					case stdout := <-out:
						if req.RequireResult {
							nowSeq++
							jobMsg := models.JobMessage{
								JobId:   job.Id,
								Seq:     nowSeq,
								Type:    models.Normal,
								Message: stdout,
							}
							if err := controller.insertJobMessage(&jobMsg); err != nil {
								return
							}
						}
					case stderr := <-errout:
						if req.RequireResult {
							nowSeq++
							jobMsg := models.JobMessage{
								JobId:   job.Id,
								Seq:     nowSeq,
								Type:    models.Error,
								Message: stderr,
							}
							if err := controller.insertJobMessage(&jobMsg); err != nil {
								return
							}
						}
					}
				}
			}()
			// https://golang.org/pkg/os/exec/#Cmd.StdoutPipe
			// https://golang.org/pkg/os/exec/#Cmd.StderrPipe
			// より
			// Wait will close the pipe after seeing the command exit, so most callers need not close the pipe themselves;
			// however, an implication is that it is incorrect to call Wait before all reads from the pipe have completed.
			// なので、標準出力と標準エラー出力を読み切ってから Wait を呼ぶ。また、標準出力と標準エラー出力の明示的な Close は不要。
			<-stdoutfin
			<-stderrfin
			if err := cmd.Wait(); err != nil {
				jobquit <- true
				if err2, ok := err.(*exec.ExitError); ok {
					job.FinishedAt = time.Now()
					job.Status = models.Finished
					if s, ok := err2.Sys().(syscall.WaitStatus); ok {
						job.ExitStatus = s.ExitStatus()
					}
				} else {
					// may be returned for I/O problems.
					job.Status = models.CannotRun
				}
				var msgCount int64
				if cerr := controller.DS.GetDB().Model(models.JobMessage{}).Where("job_id = ?", job.Id).Count(&msgCount).Error; cerr != nil {
					controller.Logger.Errorf("job_messages テーブル取得時にエラーが発生しました。error=%v", cerr)
					return
				}
				jobMsg := models.JobMessage{
					JobId:   job.Id,
					Seq:     msgCount + 1,
					Type:    models.Error,
					Message: err.Error(),
				}
				controller.insertJobMessage(&jobMsg)
			} else {
				job.FinishedAt = time.Now()
				job.Status = models.Finished
				job.ExitStatus = 0
			}
			if err := controller.DS.DoInTransaction(func(th *helpers.TxHolder) error {
				return th.GetTx().Save(&job).Error
			}); err != nil {
				controller.Logger.Errorf("jobs テーブル更新時にエラーが発生しました。error=%v", err)
			}
		}(&stdout, &stderr)

		if req.RequireResult {
			fmt.Fprintf(w, "{\"id\": %v}", job.Id)
			return
		}
	} else if !req.Async {
		cmd := exec.Command(req.Command, req.Args)
		var stdout bytes.Buffer
		var stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			if err2, ok := err.(*exec.ExitError); ok {
				if s, ok := err2.Sys().(syscall.WaitStatus); ok {
					SendEroorResponse(w, err, fmt.Sprintf("バッチが正常終了しませんでした。exitStatus=%v stdout=%v stderr=%v", s.ExitStatus(), stdout.String(), stderr.String()))
					return
				} else {
					// Unix や Winodws とは異なり、 exec.ExitError.Sys() が syscall.WaitStatus ではないOSの場合
					SendEroorResponse(w, err, fmt.Sprintf("バッチが正常終了しませんでした。stdout=%v stderr=%v", stdout.String(), stderr.String()))
					return
				}
			} else {
				// may be returned for I/O problems.
				SendEroorResponse(w, err, "バッチ実行時にエラーが発生しました。")
				return
			}
		}
		fmt.Fprintf(w, "{\"exitStatus\": 0, \"stdout\": \"%v\", \"stderr\": \"%v\"}", stdout.String(), stderr.String())
		return
	}
	encoder := json.NewEncoder(w)
	encoder.Encode(response{Error: false, Messages: []string{}})
}

func (controller *JobsController) insertJobMessage(jobMsg *models.JobMessage) error {
	if err := controller.DS.DoInTransaction(func(th *helpers.TxHolder) error {
		return th.GetTx().Create(jobMsg).Error
	}); err != nil {
		controller.Logger.Errorf("job_messages テーブル登録時にエラーが発生しました。error=%v", err)
		return err
	}
	return nil
}
