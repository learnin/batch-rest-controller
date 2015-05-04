package helpers

import (
	"encoding/json"
	"io/ioutil"
)

type Json struct {
}

func (j *Json) UnmarshalJsonFile(filePath string, dst interface{}) error {
	cfgJson, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(cfgJson, dst); err != nil {
		return err
	}
	return nil
}
