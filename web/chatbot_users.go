package web

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"../models"
)

type TemplateVarChatbotTGUserList struct {
	templateVarLayout

	Users []*models.TGUser
	Pages *templatePages
}

// HandleChatbotTGUserList displays files home
func HandleChatbotTGUserList(response http.ResponseWriter, request *http.Request) {
	start := time.Now()

	// Init Session
	tmplVars := &TemplateVarChatbotTGUserList{}
	tmpl, _ := initSessionVars(response, request, tmplVars, "templates/layout.html", "templates/chatbot_tg_user_list.html")

	// page stuff
	var entriesPerPage uint = 10

	// get Page Count
	var pageCount uint = 1
	userCount, err := models.GetTGUserCount()
	if err != nil {
		MakeErrorResponse(response, 500, err.Error(), 0)
		return
	}
	pageCount = userCount / entriesPerPage
	if userCount%entriesPerPage > 0 {
		pageCount++
	}

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
		logger.Tracef("HandleUserIndex: got 'page' query parameter: %s", pageInt)
	}

	// Add Pagination if needed
	if pageCount > 1 {
		tmplVars.Pages = makePagination("/web/admin/users/", page, pageCount, 5)
	}

	// Get Users
	users, err := models.ReadTGUserPage(entriesPerPage, page-1)
	if err != nil {
		MakeErrorResponse(response, 500, err.Error(), 0)
		return
	}

	tmplVars.Users = users

	err = tmpl.ExecuteTemplate(response, "layout", tmplVars)
	if err != nil {
		logger.Warningf("HandleChatbot: template error: %s", err.Error())
	}

	elapsed := time.Since(start)
	logger.Tracef("HandleRegistryIndex() [%s]", elapsed)
	return
}
