package web

import (
	"fmt"
	"net/http"
	"strconv"

	"../models"
)

type TemplateVarUserIndex struct {
	AlertWarn  string
	Username   string
	Users      []*models.User
	Pages      *TemplatePages
}

type TemplateVarUserNew struct {
	AlertWarn  string
	Username   string
}

func HandleUserIndex(response http.ResponseWriter, request *http.Request) {
	us, err := globalSessions.Get(request, "session-key")
	if err != nil {
		MakeErrorResponse(response, 500, err.Error(), 0)
		return
	}

	tmplVars := &TemplateVarUserIndex{}
	uid := us.Values["LoggedInUserID"].(uint)
	tmplVars.Username = models.GetUsernameByID(uid)

	// page stuff
	var entriesPerPage uint = 5

	// get Page Count
	var pageCount uint = 1
	userCount, err := models.GetUserCount()
	if err != nil {
		MakeErrorResponse(response, 500, err.Error(), 0)
		return
	}
	pageCount = userCount/entriesPerPage
	if userCount%entriesPerPage > 0 {
		pageCount++
	}
	logger.Tracef("Got %d pages.", pageCount)

	// Get Page Num
	var page uint = 1
	queryPage := request.URL.Query().Get("page")
	if queryPage != "" {
		pageInt, err := strconv.Atoi(queryPage)
		if err != nil {
			tmplVars.AlertWarn = fmt.Sprintf("Invalid page value: %s", queryPage)
		} else if pageInt < 1 || uint(pageInt) > pageCount {
			tmplVars.AlertWarn = fmt.Sprintf("Invalid page number: %d", pageInt)
		} else {
			page = uint(pageInt)
		}
		logger.Tracef("got 'page' query parameter: %s", pageInt)
	}

	// Add Pagination if needed
	if pageCount > 1 {
		tmplVars.Pages = makePagination("/web/users/", page, pageCount, 5)
	}

	// Get Users
	users, err := models.GetUsersPage(entriesPerPage, page-1)
	if err != nil {
		MakeErrorResponse(response, 500, err.Error(), 0)
		return
	}
	tmplVars.Users = users

	if err = us.Save(request, response); err != nil {
		logger.Errorf("Error saving session: %v", err)
		MakeErrorResponse(response, 500, err.Error(), 0)
		return
	}

	tmpl, err := compileTemplates("templates/layout.html", "templates/users_index.html")
	if err != nil {
		MakeErrorResponse(response, 500, err.Error(), 0)
		return
	}

	tmpl.ExecuteTemplate(response, "layout", tmplVars)
	return
}


func HandleUserNew(response http.ResponseWriter, request *http.Request) {
	us, err := globalSessions.Get(request, "session-key")
	if err != nil {
		MakeErrorResponse(response, 500, err.Error(), 0)
		return
	}

	tmplVars := &TemplateVarUserIndex{}
	uid := us.Values["LoggedInUserID"].(uint)
	tmplVars.Username = models.GetUsernameByID(uid)

	tmpl, err := compileTemplates("templates/layout.html", "templates/users_new.html")
	if err != nil {
		MakeErrorResponse(response, 500, err.Error(), 0)
		return
	}

	tmpl.ExecuteTemplate(response, "layout", tmplVars)
}