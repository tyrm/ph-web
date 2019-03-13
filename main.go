package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"./chatbot"
	"./files"
	"./models"
	"./registry"
	"./web"
	"github.com/bamzi/jobrunner"
	"github.com/gobuffalo/packr/v2"
	"github.com/gorilla/mux"
	"github.com/juju/loggo"
	"gopkg.in/alexcesaro/statsd.v2"
)

var logger *loggo.Logger
var stsd *statsd.Client
var stsdPrefix string

func main() {

	newLogger := loggo.GetLogger("main")
	logger = &newLogger

	config := CollectConfig()

	err := loggo.ConfigureLoggers(config.LoggerConfig)
	if err != nil {
		fmt.Printf("Error configurting Logger: %s", err.Error())
		return
	}

	// Connect db
	models.Init(config.DBEngine)
	defer models.Close()

	// Open StatsD Client
	stsdPrefix = config.StatsdPrefix
	sd, err := statsd.New(
		statsd.Address(config.StatsdAddress),
		)
	if err != nil {
		log.Print(err)
	}
	defer sd.Close()
	stsd = sd

	// Start Job Runner
	jobrunner.Start() // optional: jobrunner.Start(pool int, concurrent int) (10, 1)
	err = jobrunner.Schedule("@every 15s", StatusSender{})
	if err != nil {
		logger.Errorf("error starting job running: %s", err)
	}

	// Open Registry
	registry.Init(config.DBEngine, config.AESSecret)
	defer registry.Close()

	// Create Admin if no Users exist
	count, err := models.GetUserCount()
	if err != nil {
		return
	}
	if count == 0 {
		logger.Infof("Creating admin user")
		_, err := models.CreateUser("admin", "admin", "admin@admin.com")
		if err != nil {
			logger.Errorf("Error creating admin user")
		}
	}

	// Create Top Router
	web.Init(config.DBEngine, packr.New("templates", "./templates"), sd, config.StatsdPrefix, config.Debug)
	defer web.Close()

	r := mux.NewRouter()
	r.HandleFunc("/", web.HandleLanding).Methods("GET")
	r.HandleFunc("/login", web.HandleLogin)
	r.HandleFunc("/logout", web.HandleLogout)

	rWeb := r.PathPrefix("/web").Subrouter()
	rWeb.Use(web.ProtectMiddleware) // Require Valid Bearer
	rWeb.HandleFunc("/", web.HandleHome).Methods("GET")
	rWeb.HandleFunc("/chatbot/", web.HandleChatbot).Methods("GET")
	rWeb.HandleFunc("/chatbot/config", web.HandleChatbotConfig).Methods("GET")
	rWeb.HandleFunc("/chatbot/config", web.HandleChatbotConfig).Methods("POST")
	rWeb.HandleFunc("/chatbot/tg/chats/", web.HandleChatbotTGChatList).Methods("GET")
	rWeb.HandleFunc("/chatbot/tg/chats/{id}", web.HandleChatbotTGChatView).Methods("GET")
	rWeb.HandleFunc("/chatbot/tg/chats/{id}", web.HandleChatbotTGChatView).Methods("POST")
	rWeb.HandleFunc("/chatbot/tg/photos/{id}/file", web.HandleChatbotTGPhotoSizeViewFile).Methods("GET")
	rWeb.HandleFunc("/chatbot/tg/stickers/{id}/file", web.HandleChatbotTGStickerViewFile).Methods("GET")
	rWeb.HandleFunc("/chatbot/tg/users/", web.HandleChatbotTGUserList).Methods("GET")
	rWeb.HandleFunc("/files/", web.HandleFiles).Methods("GET")
	rWeb.HandleFunc("/files/config", web.HandleFilesConfig).Methods("GET")
	rWeb.HandleFunc("/files/config", web.HandleFilesConfig).Methods("POST")

	rAdmin := rWeb.PathPrefix("/admin").Subrouter()
	rAdmin.HandleFunc("/users/", web.HandleUserIndex).Methods("GET")
	rAdmin.HandleFunc("/users/", web.HandleUserIndex).Methods("POST")
	rAdmin.HandleFunc("/users/new", web.HandleUserNew).Methods("GET")
	rAdmin.HandleFunc("/users/new", web.HandleUserNew).Methods("POST")
	rAdmin.HandleFunc("/users/{id}", web.HandleUserGet).Methods("GET")
	rAdmin.HandleFunc("/registry/", web.HandleRegistryIndex).Methods("GET")
	rAdmin.HandleFunc("/registry/", web.HandleRegistryPost).Methods("POST")

	// Serve static files
	box := packr.New("staticAssets", "./static")
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(box)))

	// Serve 404
	rAdmin.PathPrefix("/").HandlerFunc(web.HandleNotFound)
	rWeb.PathPrefix("/").HandlerFunc(web.HandleNotFound)
	r.PathPrefix("/").HandlerFunc(web.HandleNotFound)

	go http.ListenAndServe(":8080", r)

	// Init Files
	go files.InitClient(false)
	go chatbot.InitClients()

	// Wait for SIGINT and SIGTERM (HIT CTRL-C)
	nch := make(chan os.Signal)
	signal.Notify(nch, syscall.SIGINT, syscall.SIGTERM)
	logger.Infof("%s", <-nch)

	logger.Infof("Done!")
}

type StatusSender struct {
}

func (_ StatusSender) Run() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	stsd.Gauge(fmt.Sprintf("%s.mem.Alloc", stsdPrefix), m.Alloc)
	stsd.Gauge(fmt.Sprintf("%s.mem.HeapAlloc", stsdPrefix), m.HeapAlloc)
	stsd.Gauge(fmt.Sprintf("%s.mem.TotalAlloc", stsdPrefix), m.TotalAlloc)
	stsd.Gauge(fmt.Sprintf("%s.mem.Sys", stsdPrefix), m.Sys)
	stsd.Gauge(fmt.Sprintf("%s.mem.NumGC", stsdPrefix), m.NumGC)
	stsd.Gauge(fmt.Sprintf("%s.goroutines", stsdPrefix), runtime.NumGoroutine())

}