package web

import (
	"fmt"
	"net/http"

	"../models"
)

// TemplateVarLanding holds template variables for HandleLanding
type TemplateVarLanding struct {
	UserID string
}

// HandleLanding displays the landing page
func HandleLanding(response http.ResponseWriter, request *http.Request) {
	defer stsd.NewTiming().Send(fmt.Sprintf("%s.web.%s.HandleLanding", stsdPrefix, request.Method))
	us, err := globalSessions.Get(request, "session-key")
	if err != nil {
		MakeErrorResponse(response, 500, err.Error(), 0)
		return
	}

	templateVars := &TemplateVarLanding{}

	if us.Values["LoggedInUserID"] != nil {
		uid := us.Values["LoggedInUserID"].(int)
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

	err = tmpl.Execute(response, templateVars)
	if err != nil {
		logger.Warningf("HandleLanding: template error: %s", err.Error())
	}
	return
}
