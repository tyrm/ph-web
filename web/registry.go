package web

import (
	"fmt"
	"net/http"
	"strings"

	"../models"
	"../registry"
)

type templateVarRegistryIndex struct {
	TemplateVarLayout

	Breadcrumbs []TemplateBreadcrumb
	Siblings    []TemplateListGroup

	ShowAddChild bool

	ModalNewChildParent string
	ModalNewChildParentID int
	ModalNewSiblingParent string
	ModalNewSiblingParentID int

	Reg         *registry.RegistryEntry
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

	// Make Breadcrumbs
	path := "/"
	if val, ok := request.URL.Query()["path"]; ok {
		path = val[0]
	}

	tmplVars.Breadcrumbs = append(tmplVars.Breadcrumbs, TemplateBreadcrumb{
		Text:   "ROOT",
		URL:    "/web/registry/?path=/",
		Active: path == "/",
	})

	paths := registry.SplitPath(path)
	startPath := "/"
	activeIndex := len(paths) - 1
	for index, key := range paths {
		tmplVars.Breadcrumbs = append(tmplVars.Breadcrumbs, TemplateBreadcrumb{
			Text:   key,
			URL:    fmt.Sprintf("/web/registry/?path=%s%s", startPath, key),
			Active: index == activeIndex,
		})
	}

	// Get Registry Entry
	reg, err := registry.GetRegistryEntry(path)
	if err != nil {
		MakeErrorResponse(response, 500, err.Error(), 0)
		return
	}
	tmplVars.Reg = reg
	logger.Tracef("got entry: %v", reg)

	// Get Children and some Template Mess
	getID := reg.ID
	tmplVars.ShowAddChild = false
	newPath := fmt.Sprintf("%s/", path)
	tmplVars.ModalNewSiblingParent = path
	if path == "/" {
		newPath = "/"
	}
	if reg.ChildCount == 0 {
		tmplVars.ShowAddChild = true

		pathLen := len(paths)
		if pathLen == 1 {
			newPath = "/"
			tmplVars.ModalNewSiblingParent = "/"
		} else {
			newPath = fmt.Sprintf("/%s/", strings.Join(paths[0:pathLen-1], "/"))
			tmplVars.ModalNewSiblingParent = fmt.Sprintf("/%s", strings.Join(paths[0:pathLen-1], "/"))
		}
		getID = reg.ParentID
		tmplVars.ModalNewSiblingParentID = reg.ParentID
	}

	tmplVars.ModalNewChildParent = path
	tmplVars.ModalNewChildParentID = reg.ID
	tmplVars.ModalNewSiblingParentID = getID

	children, err := registry.GetChildrenByID(getID)
	logger.Tracef("got children: %v", children)

	// Add Children to List
	for _, child := range children {
		tmplVars.Siblings = append(tmplVars.Siblings, TemplateListGroup{
			Text:   child.Key,
			URL:    fmt.Sprintf("/web/registry/?path=%s%s", newPath, child.Key),
			Active: reg.ID == child.ID,
			Count: child.ChildCount,
		})
	}

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
