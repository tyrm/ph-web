package web

import (
	"net/http"
	"time"

	"../registry"
)

// TemplateVarChatbotConfig holds template variables for Chatbot
type TemplateVarChatbotConfig struct {
	templateVarLayout

	IsInit bool
}

// HandleChatbotConfig displays config for chatbot
func HandleChatbotConfig(response http.ResponseWriter, request *http.Request) {
	start := time.Now()

	// Init Session
	tmplVars := &TemplateVarChatbotConfig{}
	tmpl, us := initSessionVars(response, request, tmplVars, "templates/layout.html", "templates/chatbot_config.html")

	if request.Method == "POST" {
		uid := us.Values["LoggedInUserID"].(int)

		err := request.ParseForm()
		if err != nil {
			MakeErrorResponse(response, 500, err.Error(), 0)
			return
		}
		logger.Tracef("got post: %v", request.Form)

		formAction := ""
		if val, ok := request.Form["_action"]; ok {
			formAction = val[0]
		} else {
			MakeErrorResponse(response, 400, "missing action", 0)
			return
		}

		if formAction == "config_telegram" {
			formToken := ""
			if val, ok := request.Form["token"]; ok {
				formToken = val[0]
			}

			if formToken != "" {
				tmplVars.AlertSuccess = "Telegram token updated."
				_, err := registry.Set("/system/chatbot/telegram/token", formToken, true, uid)
				if err != nil {
					MakeErrorResponse(response, 500, err.Error(), 0)
					return
				}
			} else {
				tmplVars.AlertError = "Missing telegram token."
			}
		}
	}

	elapsed := time.Since(start)
	tmplVars.DebugTime = elapsed.String()
	err := tmpl.ExecuteTemplate(response, "layout", tmplVars)
	if err != nil {
		logger.Warningf("HandleChatbotConfig: template error: %s", err.Error())
	}

	elapsed = time.Since(start)
	logger.Tracef("HandleChatbotConfig() [%s]", elapsed)
	return
}