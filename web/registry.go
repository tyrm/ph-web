package web

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"../registry"
)

// templateVarRegistryIndex holds template variables for HandleRegistryIndex
type templateVarRegistryIndex struct {
	templateVarLayout

	Breadcrumbs []templateBreadcrumb
	Siblings    []templateListGroup

	DisableAddChild bool
	DisableDelete   bool

	ModalNewChildParent     string
	ModalNewChildParentID   int
	ModalNewSiblingParent   string
	ModalNewSiblingParentID int

	Reg *registry.Entry
}

// HandleRegistryPost handles registry change requests
func HandleRegistryPost(response http.ResponseWriter, request *http.Request) {
	defer stsd.NewTiming().Send(fmt.Sprintf("%s.web.%s.HandleRegistryPost", stsdPrefix, request.Method))
	start := time.Now()

	// Init Session
	_, us := initSession(response, request)
	uid := us.Values["LoggedInUserID"].(int)

	err := request.ParseForm()
	if err != nil {
		MakeErrorResponse(response, 500, err.Error(), 0)
		return
	}

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
			if err != nil {
				MakeErrorResponse(response, 400, fmt.Sprintf("invalid registry id: %s", val[0]), 0)
				return
			}

			formRegID = i
		} else {
			MakeErrorResponse(response, 400, "missing registry id", 0)
			return
		}

		entry, err := registry.GetByID(formRegID)
		if err != nil {
			MakeErrorResponse(response, 500, err.Error(), 0)
			return
		}

		// don't delete root
		if entry.Key == "{ROOT}" {
			MakeErrorResponse(response, 403, "Can't delete root", 0)
		}

		parentID := entry.ParentID
		parentPath, err := registry.GetPathByID(parentID)
		if err != nil {
			MakeErrorResponse(response, 500, err.Error(), 0)
			return
		}

		err = entry.Delete()
		if err != nil {
			MakeErrorResponse(response, 500, err.Error(), 0)
			return
		}

		newPage := fmt.Sprintf("/web/admin/registry/?path=%s", parentPath)
		response.Header().Set("Location", newPage)
		response.WriteHeader(http.StatusFound)
		return
	} else if formAction == "create" {
		logger.Debugf("got create")

		// Gather form Variables
		formParentID := 0
		if val, ok := request.Form["parent_id"]; ok {
			i, err := strconv.Atoi(val[0])
			if err != nil {
				MakeErrorResponse(response, 400, fmt.Sprintf("invalid registry id: %s", val[0]), 0)
				return
			}

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
		newRegEntry, err := registry.New(formParentID, formKey, formValue, formSecure, uid)
		if err != nil {
			MakeErrorResponse(response, 500, err.Error(), 0);
			return
		}

		// Redirect to new path
		parentPath, err := registry.GetPathByID(newRegEntry.ID)
		newPage := fmt.Sprintf("/web/admin/registry/?path=%s", parentPath)
		response.Header().Set("Location", newPage)
		response.WriteHeader(http.StatusFound)
		return
	} else if formAction == "update" {
		// Gather form Variables
		formRegID := 0
		if val, ok := request.Form["reg_id"]; ok {
			i, err := strconv.Atoi(val[0])
			if err != nil {
				MakeErrorResponse(response, 400, fmt.Sprintf("invalid registry id: %s", val[0]), 0)
				return
			}
			formRegID = i
		} else {
			MakeErrorResponse(response, 400, "missing registry id", 0)
			return
		}

		formValue := ""
		if val, ok := request.Form["value"]; ok {
			formValue = val[0]
		} else {
			MakeErrorResponse(response, 400, "missing registry value", 0)
			return
		}

		reg, err := registry.GetByID(formRegID)
		if err != nil {
			MakeErrorResponse(response, 500, err.Error(), 0)
			return
		}

		err = reg.SetValue(formValue)
		if err != nil {
			MakeErrorResponse(response, 500, err.Error(), 0)
			return
		}

		// Redirect to new path
		parentPath, err := registry.GetPathByID(reg.ID)
		newPage := fmt.Sprintf("/web/admin/registry/?path=%s", parentPath)
		response.Header().Set("Location", newPage)
		response.WriteHeader(http.StatusFound)
		return

	}

	MakeErrorResponse(response, 400, fmt.Sprintf("unknown action: %s", formAction), 0)

	elapsed := time.Since(start)
	logger.Tracef("HandleRegistryPost() [%s]", elapsed)
	return
}

// HandleRegistryIndex displays registry tree
func HandleRegistryIndex(response http.ResponseWriter, request *http.Request) {
	defer stsd.NewTiming().Send(fmt.Sprintf("%s.web.%s.HandleRegistryIndex", stsdPrefix, request.Method))
	start := time.Now()

	// Init Session
	tmplVars := &templateVarRegistryIndex{}
	tmpl, _ := initSessionVars(response, request, tmplVars, "templates/layout.html", "templates/registry.html")

	path := "/"
	if val, ok := request.URL.Query()["path"]; ok {
		path = val[0]
	}
	paths := registry.SplitPath(path)

	// Get Registry Entry
	reg, err := registry.Get(path)
	if err != nil {
		MakeErrorResponse(response, 500, err.Error(), 0);
		return
	}
	tmplVars.Reg = reg

	// Get Children and some Template Mess
	getID := reg.ID
	tmplVars.DisableAddChild = true
	tmplVars.DisableDelete = true
	newPath := fmt.Sprintf("%s/", path)
	tmplVars.ModalNewSiblingParent = path

	if path == "/" {
		newPath = "/"
	}
	if reg.ChildCount == 0 {

		pathLen := len(paths)
		if pathLen == 1 {
			newPath = "/"
			tmplVars.ModalNewSiblingParent = "/"
		} else {
			newPath = fmt.Sprintf("/%s/", strings.Join(paths[0:pathLen-1], "/"))
			tmplVars.ModalNewSiblingParent = fmt.Sprintf("/%s", strings.Join(paths[0:pathLen-1], "/"))
		}
		tmplVars.DisableDelete = false
		tmplVars.DisableAddChild = false
		getID = reg.ParentID
	}
	if path == "/" {
		tmplVars.DisableDelete = true
		tmplVars.DisableAddChild = true
		tmplVars.ModalNewSiblingParentID = reg.ID
	} else {
		tmplVars.ModalNewSiblingParentID = getID
	}

	tmplVars.ModalNewChildParent = path
	tmplVars.ModalNewChildParentID = reg.ID

	children, err := registry.GetChildrenByID(getID)

	// Make Breadcrumbs
	tmplVars.Breadcrumbs = append(tmplVars.Breadcrumbs, templateBreadcrumb{
		Text:   "ROOT",
		URL:    "/web/admin/registry/?path=/",
		Active: path == "/",
	})

	startPath := "/"
	activeIndex := len(paths) - 1
	for index, key := range paths {
		newKey := key
		if reg.ChildCount == 0 && index == activeIndex {
			newKey = ""
		}

		tmplVars.Breadcrumbs = append(tmplVars.Breadcrumbs, templateBreadcrumb{
			Text:   newKey,
			URL:    fmt.Sprintf("/web/admin/registry/?path=%s%s", startPath, key),
			Active: index == activeIndex,
		})
		startPath = startPath + key + "/"
	}
	if reg.ChildCount != 0 && path != "/" {
		tmplVars.Breadcrumbs = append(tmplVars.Breadcrumbs, templateBreadcrumb{
			Text:   "",
			URL:    "#",
			Active: true,
		})
	}

	// Add Children to List
	for _, child := range children {
		icon := ""
		if child.Secure {
			icon = "lock"
		}

		tmplVars.Siblings = append(tmplVars.Siblings, templateListGroup{
			Text:    child.Key,
			URL:     fmt.Sprintf("/web/admin/registry/?path=%s%s", newPath, child.Key),
			Active:  reg.ID == child.ID,
			Count:   child.ChildCount,
			FAIconR: icon,
		})
	}

	elapsed := time.Since(start)
	tmplVars.DebugTime = elapsed.String()
	err = tmpl.ExecuteTemplate(response, "layout", tmplVars)
	if err != nil {
		logger.Warningf("HandleRegistryIndex: template error: %s", err.Error())
	}

	elapsed = time.Since(start)
	logger.Tracef("HandleRegistryIndex() [%s]", elapsed)
	return
}
