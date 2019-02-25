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
	initSessionVars(response, request, tmplVars)

	tmpl, err := compileTemplates("templates/layout.html", "templates/chatbot.html")
	if err != nil {
		MakeErrorResponse(response, 500, err.Error(), 0)
		return
	}

	err = tmpl.ExecuteTemplate(response, "layout", tmplVars)
	if err != nil {
		logger.Warningf("Error executing template: %s", err)
	}

	elapsed := time.Since(start)
	logger.Tracef("HandleRegistryIndex() [%s]", elapsed)
	return
}