package deepmock

import (
	"regexp"
	"strings"
)

func (rm *RequestMatch) Match(method, path string) bool {
	rm.once.Do(func() {
		rm.Method = strings.ToUpper(rm.Method)
		rm.re = regexp.MustCompile(rm.Path)
	})

	if rm.Method != method {
		return false
	}
	return rm.re.MatchString(path)
}
