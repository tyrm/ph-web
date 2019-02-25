package web

import (
	"net/http"
)

// TemplateVarHome holds template variables for HandleHome
type TemplateVarHome struct {
	templateVarLayout
}

// HandleHome displays the home dashboard
func HandleHome(response http.ResponseWriter, request *http.Request) {
	// Init Session
	tmplVars := &TemplateVarHome{}
	tmpl, _ := initSessionVars(response, request, tmplVars, "templates/layout.html", "templates/home.html")

	err := tmpl.ExecuteTemplate(response, "layout", tmplVars)
	if err != nil {
		logger.Warningf("HandleHome: template error: %s", err.Error())
	}
	return
}
