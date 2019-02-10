package web

import (
	"net/http"

	"../models"
	)


type TemplateVarHome struct {
	AlertWarn  string
	Username   string
}

func HandleHome(response http.ResponseWriter, request *http.Request) {
	us, err := globalSessions.Get(request, "session-key")
	if err != nil {
		MakeErrorResponse(response, 500, err.Error(), 0)
		return
	}

	tmplVars := &TemplateVarHome{}
	uid := us.Values["LoggedInUserID"].(uint)
	tmplVars.Username = models.GetUsernameByID(uid)

	tmpl, err := compileTemplates("templates/layout.html", "templates/home.html")
	if err != nil {
		MakeErrorResponse(response, 500, err.Error(), 0)
		return
	}

	tmpl.ExecuteTemplate(response, "layout", tmplVars)
}
