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
)

var logger *loggo.Logger

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

	// Create Admin if no Users exist
	count, err := models.EstimateCountUsers()
	if err != nil {
		return
	}
	if count == 0 {
		logger.Infof("Creating admin user")
		_, err := models.NewUser("admin", "admin")
		if err != nil {
			logger.Errorf("Error creating admin user")
		}
	}

	web.Init(config.DBEngine)
	defer web.Close()

	// Create Top Router
	r := mux.NewRouter()
	r.HandleFunc("/", web.HandleLanding).Methods("GET")
	r.HandleFunc("/login", web.HandleLogin)
	r.HandleFunc("/logout", web.HandleLogout)

	rWeb := r.PathPrefix("/web").Subrouter()
	rWeb.Use(web.ProtectMiddleware) // Require Valid Bearer
	rWeb.HandleFunc("/users", web.HandleUserIndex).Methods("GET")

	// Serve static files
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(box)))

	// Serve 404
	rWeb.PathPrefix("/").HandlerFunc(web.HandleNotFound)
	r.PathPrefix("/").HandlerFunc(web.HandleNotFound)

	go http.ListenAndServe(":8080", r)

	// Wait for SIGINT and SIGTERM (HIT CTRL-C)
	nch := make(chan os.Signal)
	signal.Notify(nch, syscall.SIGINT, syscall.SIGTERM)
	logger.Infof("%s", <-nch)

	logger.Infof("Done!")
}
