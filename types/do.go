package types

import "time"

type (
	RuleDO struct {
		ID        string    `ddb:"id"`
		Path      string    `ddb:"path"`
		Method    string    `ddb:"method"`
		Variable  []byte    `ddb:"variable"`
		Weight    []byte    `ddb:"weight"`
		Responses []byte    `ddb:"responses"`
		Version   int       `ddb:"version"`
		CTime     time.Time `ddb:"ctime"`
		MTime     time.Time `ddb:"mtime"`
		Disabled  bool      `ddb:"disabled"`
	}
)
