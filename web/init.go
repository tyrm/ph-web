package web

import (
	"github.com/antonlindstrom/pgstore"
	"github.com/gobuffalo/packr/v2"
	"github.com/juju/loggo"
)

var logger *loggo.Logger
var templates *packr.Box

var globalSessions *pgstore.PGStore
func Close() {
	globalSessions.Close()
}

func Init(db string) {
	newLogger := loggo.GetLogger("web.web")
	logger = &newLogger

	gs, err := pgstore.NewPGStore(db, []byte("secret-key"))
	if err != nil {
		logger.Errorf(err.Error())
	}
	globalSessions = gs


	templates = packr.New("templates", "./templates")
}
