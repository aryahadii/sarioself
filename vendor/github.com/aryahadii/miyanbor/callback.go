package miyanbor

import (
	"regexp"
)

// CallbackFunction will be called to push new update
type CallbackFunction func(*UserSession, []string, interface{})

// callback contains a CallbackFunction and it's related Regex pattern
type callback struct {
	Pattern  *regexp.Regexp
	Function CallbackFunction
}
