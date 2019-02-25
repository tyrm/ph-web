package web

import (
	"net/http"
	"time"

	"../chatbot/telegram"
)

// TemplateVarFiles holds template variables for HandleFiles
type TemplateVarChatbot struct {
	templateVarLayout

	IsInit bool
}

// TelegramIsInit returns true if telegram is connected
func (_ *TemplateVarChatbot) TelegramIsInit() bool {
	return telegram.IsInit()
}

// HandleChatbot displays files home
func HandleChatbot(response http.ResponseWriter, request *http.Request) {
	start := time.Now()

	// Init Session
	tmplVars := &TemplateVarChatbot{}
	tmpl, _ := initSessionVars(response, request, tmplVars, "templates/layout.html", "templates/chatbot.html")

	err := tmpl.ExecuteTemplate(response, "layout", tmplVars)
	if err != nil {
		logger.Warningf("HandleChatbot: template error: %s", err.Error())
	}

	elapsed := time.Since(start)
	logger.Tracef("HandleRegistryIndex() [%s]", elapsed)
	return
}