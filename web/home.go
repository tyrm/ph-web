package web

import (
	"net/http"

	"../models"
)

type TemplateVarHome struct {
	TemplateVarLayout
}

func HandleHome(response http.ResponseWriter, request *http.Request) {
	us, err := globalSessions.Get(request, "session-key")
	if err != nil {
		MakeErrorResponse(response, 500, err.Error(), 0)
		return
	}

	tmplVars := &TemplateVarHome{}
	uid := us.Values["LoggedInUserID"].(string)
	tmplVars.Username = models.GetUsernameByID(uid)
	tmplVars.NavBar = makeNavbar(request.URL.Path)

	tmpl, err := compileTemplates("templates/layout.html", "templates/home.html")
	if err != nil {
		MakeErrorResponse(response, 500, err.Error(), 0)
		return
	}

	tmpl.ExecuteTemplate(response, "layout", tmplVars)
}
