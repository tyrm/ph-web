package web

import "github.com/juju/loggo"

var logger *loggo.Logger

func init() {
	newLogger := loggo.GetLogger("puphaus.web")
	logger = &newLogger
}
