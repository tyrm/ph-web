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
	initSessionVars(response, request, tmplVars)

	tmpl, err := compileTemplates("templates/layout.html", "templates/home.html")
	if err != nil {
		MakeErrorResponse(response, 500, err.Error(), 0)
		return
	}

	tmpl.ExecuteTemplate(response, "layout", tmplVars)
}
