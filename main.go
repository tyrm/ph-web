package main

import (
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"./models"
	"./web"
	"github.com/gobuffalo/packr/v2"
	"github.com/gorilla/mux"
	"github.com/juju/loggo"
	"golang.org/x/crypto/bcrypt"
)

var logger *loggo.Logger

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func main() {
	loggo.ConfigureLoggers("<root>=TRACE")

	newLogger := loggo.GetLogger("web")
	logger = &newLogger

	config := CollectConfig()

	// Load Static Assets
	box := packr.New("staticAssets", "./static")

	// Connect DB
	models.InitDB(config.DBEngine)
	defer models.CloseDB()

	// Create Top Router
	r := mux.NewRouter()
	r.HandleFunc("/", web.HandleLanding).Methods("GET")

	// Serve static files
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(box)))

	// Serve 404
	r.PathPrefix("/").HandlerFunc(web.HandleNotFound)

	go http.ListenAndServe(":8080", r)

	// Wait for SIGINT and SIGTERM (HIT CTRL-C)
	nch := make(chan os.Signal)
	signal.Notify(nch, syscall.SIGINT, syscall.SIGTERM)
	logger.Infof("%s", <-nch)

	logger.Infof("Done!")
}
