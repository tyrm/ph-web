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
	tmpl, _ := initSessionVars(response, request, tmplVars, "templates/layout.html", "templates/chatbot_config.html")

	err := tmpl.ExecuteTemplate(response, "layout", tmplVars)
	if err != nil {
		logger.Warningf("HandleChatbotConfig: template error: %s", err.Error())
	}

	elapsed := time.Since(start)
	logger.Tracef("HandleChatbotConfig() [%s]", elapsed)
	return
}

// HandleChatbotConfig displays config for chatbot
func HandleChatbotConfigPost(response http.ResponseWriter, request *http.Request) {
	start := time.Now()

	// Init Session
	_, us := initSession(response, request)
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

		// Get Parent or Create
		var regParent *registry.Entry
		regParent, err = registry.Get("/system/files")
		if err != nil {
			logger.Errorf("Error getting /system/files: %s", err.Error())
			if err == registry.ErrDoesNotExist {
				logger.Infof("Could not get /system/files, creating")
				var regSystem *registry.Entry
				regSystem, err2 := registry.Get("/system")
				if err2 != nil {
					if err == registry.ErrDoesNotExist {
						logger.Infof("Could not get /system, creating")
						var regRoot *registry.Entry
						regRoot, err3 := registry.Get("/")
						if err3 != nil {
							logger.Errorf("Could not get root: %s", err3.Error())
							MakeErrorResponse(response, 500, err.Error(), 0)
							return
						}
						var errNew error
						regSystem, errNew = registry.New(regRoot.ID, "system", "", false, uid)
						if errNew != nil {
							logger.Errorf("Could not create /system/files: %s", errNew.Error())
							MakeErrorResponse(response, 500, err.Error(), 0)
							return
						}

					} else {
						logger.Errorf("Could not get /system: %s", err2.Error())
						MakeErrorResponse(response, 500, err.Error(), 0)
						return
					}
				}
				var errNew error
				regParent, errNew = registry.New(regSystem.ID, "files", "", false, uid)
				if errNew != nil {
					logger.Errorf("Could not create /system/files: %s", errNew.Error())
					MakeErrorResponse(response, 500, err.Error(), 0)
					return
				}
			} else {
				logger.Errorf("Could not get /system: %s", err.Error())
				MakeErrorResponse(response, 500, err.Error(), 0)
				return
			}
		}

	}

	elapsed := time.Since(start)
	logger.Tracef("HandleChatbotConfig() [%s]", elapsed)
	return
}