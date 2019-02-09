package web

import (
	"net/http"

	"../models"
)

type TemplateVarLanding struct {
	UserID  string
}

func HandleLanding(response http.ResponseWriter, request *http.Request) {
	us, err := globalSessions.Get(request, "session-key")
	if err != nil {
		MakeErrorResponse(response, 500, err.Error(), 0)
		return
	}

	templateVars := &TemplateVarLanding{}

	if us.Values["LoggedInUserID"] != nil {
		uid := us.Values["LoggedInUserID"].(uint)
		templateVars.UserID = models.GetUsernameByID(uid)
	}

	if err = us.Save(request, response); err != nil {
		logger.Errorf("Error saving session: %v", err)
		MakeErrorResponse(response, 500, err.Error(), 0)
		return
	}

	tmpl, err := compileTemplates("templates/landing.html")
	if err != nil {
		MakeErrorResponse(response, 500, err.Error(), 0)
		return
	}

	tmpl.Execute(response, templateVars)
	return
}