package web

import (
	"net/http"
	"../models"
)
type templateVarRegistryIndex struct {
	TemplateVarLayout

	Breadcrumbs []TemplateBreadcrumb
}

func HandleRegistryIndex(response http.ResponseWriter, request *http.Request) {
	// Init Session
	us, err := globalSessions.Get(request, "session-key")
	if err != nil {
		MakeErrorResponse(response, 500, err.Error(), 0)
		return
	}

	tmplVars := &templateVarRegistryIndex{}
	uid := us.Values["LoggedInUserID"].(string)
	tmplVars.Username = models.GetUsernameByID(uid)
	tmplVars.NavBar = makeNavbar(request.URL.Path)

	// Do Stuff
	path := "/"
	if val, ok := request.URL.Query()["path"]; ok {
		path = val[0]
	}
	logger.Tracef("got path: %s", path)

	tmplVars.Breadcrumbs = append(tmplVars.Breadcrumbs, TemplateBreadcrumb{
		Text: "ROOT",
		URL: "/web/registry/?path=/",
	})
	tmplVars.Breadcrumbs = append(tmplVars.Breadcrumbs, TemplateBreadcrumb{
		Text: "asdf",
		URL: "/web/registry/?path=/asdf",
		Active: true,
	})

	// Compile Template
	tmpl, err := compileTemplates("templates/layout.html", "templates/registry_index.html")
	if err != nil {
		MakeErrorResponse(response, 500, err.Error(), 0)
		return
	}

	err = tmpl.ExecuteTemplate(response, "layout", tmplVars)
	if err != nil {
		logger.Errorf("HandleUserIndex: Error executing template: %v", err)
	}
	return
}
