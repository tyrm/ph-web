package main

import (
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"./models"
	"./registry"
	"./web"
	"github.com/gobuffalo/packr/v2"
	"github.com/gorilla/mux"
	"github.com/juju/loggo"
)

var logger *loggo.Logger

func main() {
	loggo.ConfigureLoggers("<root>=TRACE")

	newLogger := loggo.GetLogger("main")
	logger = &newLogger

	config := CollectConfig()

	// Connect db
	models.Init(config.DBEngine)
	defer models.CloseDB()

	// Open Registry
	registry.Init(config.DBEngine, config.AESSecret)
	defer registry.Close()

	registry.GetPathByID(5)

	// Create Admin if no Users exist
	count, err := models.EstimateCountUsers()
	if err != nil {
		return
	}
	if count == 0 {
		logger.Infof("Creating admin user")
		_, err := models.NewUser("admin", "admin", "admin@admin.com")
		if err != nil {
			logger.Errorf("Error creating admin user")
		}
	}

	// Create Top Router
	web.Init(config.DBEngine, packr.New("templates", "./templates"))
	defer web.Close()

	r := mux.NewRouter()
	r.HandleFunc("/", web.HandleLanding).Methods("GET")
	r.HandleFunc("/login", web.HandleLogin)
	r.HandleFunc("/logout", web.HandleLogout)

	rWeb := r.PathPrefix("/web").Subrouter()
	rWeb.Use(web.ProtectMiddleware) // Require Valid Bearer
	rWeb.HandleFunc("/", web.HandleHome).Methods("GET")
	rWeb.HandleFunc("/users/", web.HandleUserIndex).Methods("GET")
	rWeb.HandleFunc("/users/new", web.HandleUserNew).Methods("GET")
	rWeb.HandleFunc("/users/new", web.HandleUserNew).Methods("POST")
	rWeb.HandleFunc("/users/{id}", web.HandleUserGet).Methods("GET")
	rWeb.HandleFunc("/registry/", web.HandleRegistryIndex).Methods("GET")
	rWeb.HandleFunc("/registry/", web.HandleRegistryPost).Methods("POST")

	// Serve static files
	box := packr.New("staticAssets", "./static")
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
