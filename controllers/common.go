package controllers

import (
	"encoding/json"
	"net/http"
)

const DEBUG = true

type response struct {
	Error        bool     `json:"error"`
	Messages     []string `json:"messages"`
	DebugMessage string   `json:"debugMessage"`
}

func sendEroorResponse(w http.ResponseWriter, e error, messages ...string) {
	if messages[0] == "" {
		messages = []string{"システムエラーが発生しました。"}
	}
	res := response{
		Error:    true,
		Messages: messages,
	}
	if DEBUG && e != nil {
		res.DebugMessage = e.Error()
	}
	encoder := json.NewEncoder(w)
	encoder.Encode(res)
}
