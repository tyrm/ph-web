package web

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

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

func HandleRegistryPost(response http.ResponseWriter, request *http.Request) {
	start := time.Now()

	// Init Session
	_, err := globalSessions.Get(request, "session-key")
	if err != nil {
		MakeErrorResponse(response, 500, err.Error(), 0)
		return
	}

	err = request.ParseForm()
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

	if formAction == "delete" {
		logger.Debugf("got delete")

		formRegID := 0
		if val, ok := request.Form["reg_id"]; ok {
			i, err := strconv.Atoi(val[0])
			if err != nil {MakeErrorResponse(response, 400, fmt.Sprintf("invalid registry id: %s", val[0]), 0); return}

			formRegID = i
		} else {
			MakeErrorResponse(response, 400, "missing registry id", 0)
			return
		}

		entry, err := registry.GetByID(formRegID)
		if err != nil {MakeErrorResponse(response, 500, err.Error(), 0); return}

		parentID := entry.ParentID
		parentPath, err := registry.GetPathByID(parentID)
		if err != nil {MakeErrorResponse(response, 500, err.Error(), 0); return}

		err = entry.Delete()
		if err != nil {MakeErrorResponse(response, 500, err.Error(), 0); return}

		newPage := fmt.Sprintf("/web/registry/?path=%s", parentPath)
		response.Header().Set("Location", newPage)
		response.WriteHeader(http.StatusFound)
		return
	} else if formAction == "create" {
		logger.Debugf("got create")

		// Gather form Variables
		formParentID := 0
		if val, ok := request.Form["parent_id"]; ok {
			i, err := strconv.Atoi(val[0])
			if err != nil {MakeErrorResponse(response, 400, fmt.Sprintf("invalid registry id: %s", val[0]), 0); return}

			formParentID = i
		} else {
			MakeErrorResponse(response, 400, "missing registry id", 0)
			return
		}

		formKey := ""
		if val, ok := request.Form["key"]; ok {
			formKey = val[0]
		} else {
			MakeErrorResponse(response, 400, "missing registry key", 0)
			return
		}

		formValue := ""
		if val, ok := request.Form["value"]; ok {
			formValue = val[0]
		} else {
			MakeErrorResponse(response, 400, "missing registry value", 0)
			return
		}

		formSecure := false
		if val, ok := request.Form["secure"]; ok {
			if val[0] == "true" {
				formSecure = true
			}
		}

		// Create record
		newRegEntry, err := registry.New(formParentID, formKey, formValue, formSecure)
		if err != nil {MakeErrorResponse(response, 500, err.Error(), 0); return}

		// Redirect to new path
		parentPath, err := registry.GetPathByID(newRegEntry.ID)
		newPage := fmt.Sprintf("/web/registry/?path=%s", parentPath)
		response.Header().Set("Location", newPage)
		response.WriteHeader(http.StatusFound)
		return
	}

	MakeErrorResponse(response, 400, fmt.Sprintf("unknown action: %s", formAction), 0)

	elapsed := time.Since(start)
	logger.Tracef("HandleRegistryPost() [%s]", elapsed)
	return
}

func HandleRegistryIndex(response http.ResponseWriter, request *http.Request) {
	start := time.Now()

	// Init Session
	us, err := globalSessions.Get(request, "session-key")
	if err != nil {MakeErrorResponse(response, 500, err.Error(), 0); return}

	tmplVars := &templateVarRegistryIndex{}
	uid := us.Values["LoggedInUserID"].(string)
	tmplVars.Username = models.GetUsernameByID(uid)
	tmplVars.NavBar = makeNavbar(request.URL.Path)

	path := "/"
	if val, ok := request.URL.Query()["path"]; ok {
		path = val[0]
	}
	paths := registry.SplitPath(path)

	// Get Registry Entry
	reg, err := registry.Get(path)
	if err != nil {MakeErrorResponse(response, 500, err.Error(), 0); return}
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


	// Make Breadcrumbs
	tmplVars.Breadcrumbs = append(tmplVars.Breadcrumbs, TemplateBreadcrumb{
		Text:   "ROOT",
		URL:    "/web/registry/?path=/",
		Active: path == "/",
	})

	startPath := "/"
	activeIndex := len(paths) - 1
	for index, key := range paths {
		newKey := key
		if reg.ChildCount == 0 && index == activeIndex {
			newKey = ""
		}

		tmplVars.Breadcrumbs = append(tmplVars.Breadcrumbs, TemplateBreadcrumb{
			Text:   newKey,
			URL:    fmt.Sprintf("/web/registry/?path=%s%s", startPath, key),
			Active: index == activeIndex,
		})
		startPath = startPath + key + "/"
	}
	if reg.ChildCount != 0 && path != "/" {
		tmplVars.Breadcrumbs = append(tmplVars.Breadcrumbs, TemplateBreadcrumb{
			Text:   "",
			URL:    "#",
			Active: true,
		})
	}

	// Add Children to List
	for _, child := range children {
		icon := ""
		if child.Secure{
			icon = "lock"
		}

		tmplVars.Siblings = append(tmplVars.Siblings, TemplateListGroup{
			Text:   child.Key,
			URL:    fmt.Sprintf("/web/registry/?path=%s%s", newPath, child.Key),
			Active: reg.ID == child.ID,
			Count: child.ChildCount,
			FAIconR: icon,
		})
	}

	// Compile Template
	tmpl, err := compileTemplates("templates/layout.html", "templates/registry_index.html")
	if err != nil {MakeErrorResponse(response, 500, err.Error(), 0); return}

	err = tmpl.ExecuteTemplate(response, "layout", tmplVars)
	if err != nil {logger.Errorf("HandleUserIndex: Error executing template: %v", err)}

	elapsed := time.Since(start)
	logger.Tracef("HandleRegistryIndex() [%s]", elapsed)
	return
}
