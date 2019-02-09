package web

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"../models"
)

type TemplateVarUserIndex struct {
	AlertWarn  string
	Users      []*models.User
	Pages      *TemplatePages
}

func HandleUserIndex(response http.ResponseWriter, request *http.Request) {
	us, err := globalSessions.Get(request, "session-key")

	tmplVars := &TemplateVarUserIndex{}

	if err != nil {
		MakeErrorResponse(response, 500, err.Error(), 0)
		return
	}

	tmlpLayoutStr, err := templates.FindString("templates/layout.html")
	if err != nil {
		MakeErrorResponse(response, 500, err.Error(), 0)
		return
	}
	tmlpUserIndexStr, err := templates.FindString("templates/users_index.html")
	if err != nil {
		MakeErrorResponse(response, 500, err.Error(), 0)
		return
	}

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

	//
	if pageCount > 1 {
		tmplVars.Pages = makePagination("/web/users", page, pageCount, 5)
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

	tmpl := template.New("landing template")
	tmpl = template.Must(tmpl.Parse(tmlpUserIndexStr + tmlpLayoutStr))

	tmpl.ExecuteTemplate(response, "layout", tmplVars)
	return
}
