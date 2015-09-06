package models

import (
	"time"
)

//go:generate stringer -type=JobStatus
type JobStatus int

const (
	WaitingToRun JobStatus = iota + 1
	Running
	Finished
	CannotRun
)

type Job struct {
	Id         int64     `sql:"AUTO_INCREMENT"`
	Command    string    `sql:"size:1000;not null"`
	Args       string    `sql:"size:1000"`
	Status     JobStatus `sql:"size:1;not null"`
	ExitStatus int       `sql:"size:4`
	CreatedAt  time.Time `sql:"DEFAULT:current_timestamp;not null"`
	FinishedAt time.Time
}

//go:generate stringer -type=JobMessageType
type JobMessageType int

const (
	Normal JobMessageType = iota + 1
	Error
)

type JobMessage struct {
	JobId     int64          `gorm:"primary_key" sql:"type:bigint"`
	Seq       int64          `gorm:"primary_key" sql:"type:bigint"`
	Type      JobMessageType `sql:"size:1;not null"`
	Message   string         `sql:"size:4000"`
	CreatedAt time.Time      `sql:"DEFAULT:current_timestamp;not null"`
}

type ApiKey struct {
	Id         int64     `sql:"AUTO_INCREMENT"`
	ClientName string    `sql:"size:100;not null"`
	ApiKey     string    `sql:"size:128;not null;unique_index"`
	CreatedAt  time.Time `sql:"DEFAULT:current_timestamp;not null"`
}
