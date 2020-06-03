package entity

import (
	"time"
)

type Rule struct {
	ID           string    `ddb:"id"`
	Path         string    `ddb:"path"`
	Method       string    `ddb:"method"`
	Variable     []byte    `ddb:"variable"`
	Weight       []byte    `ddb:"weight"`
	Responses    []byte    `ddb:"responses"`
	Version      int       `ddb:"version"`
	CreatedTime  time.Time `ddb:"ctime"`
	ModifiedTime time.Time `ddb:"mtime"`
	Disabled     bool      `ddb:"disabled"`
}
