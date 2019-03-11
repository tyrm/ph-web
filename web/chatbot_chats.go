package web

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"../models"
	"github.com/gorilla/mux"
)

type TemplateVarChatbotTGChatList struct {
	templateVarLayout

	Chats []*models.TGChat
	Pages *templatePages
}

type TemplateVarChatbotTGChatView struct {
	templateVarLayout

	TGChat *models.TGChat

	IsInit bool
}

// HandleChatbot displays files home
func HandleChatbotTGChatList(response http.ResponseWriter, request *http.Request) {
	start := time.Now()

	// Init Session
	tmplVars := &TemplateVarChatbotTGChatList{}
	tmpl, _ := initSessionVars(response, request, tmplVars, "templates/layout.html", "templates/chatbot_tg_chat_list.html")

	// page stuff
	var entriesPerPage uint = 10

	// get Page Count
	var pageCount uint = 1
	userCount, err := models.GetTGChatCount()
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
	chats, err := models.ReadTGChatPage(entriesPerPage, page-1)
	if err != nil {
		MakeErrorResponse(response, 500, err.Error(), 0)
		return
	}
	tmplVars.Chats = chats

	elapsed := time.Since(start)
	tmplVars.DebugTime = elapsed.String()
	err = tmpl.ExecuteTemplate(response, "layout", tmplVars)
	if err != nil {
		logger.Warningf("HandleChatbot: template error: %s", err.Error())
	}

	elapsed = time.Since(start)
	logger.Tracef("HandleRegistryIndex() [%s]", elapsed)
	return
}

// HandleChatbot displays files home
func HandleChatbotTGChatView(response http.ResponseWriter, request *http.Request) {
	start := time.Now()

	// Init Session
	tmplVars := &TemplateVarChatbotTGChatView{}
	tmpl, _ := initSessionVars(response, request, tmplVars, "templates/layout.html", "templates/chatbot_tg_chat_view.html")

	vars := mux.Vars(request)
	n, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		MakeErrorResponse(response, 500, err.Error(), 0)
		logger.Errorf("HandleChatbotTGChatView: Error getting chat: %v", err)
		return
	}

	chat, err := models.ReadTGChatByAPIID(n)
	if err != nil {
		if err == models.ErrDoesNotExist {
			tmplVars.AlertWarn = fmt.Sprintf("Chat [%s] doesn't exist.", vars["id"])
		} else {
			MakeErrorResponse(response, 500, err.Error(), 0)
			logger.Errorf("HandleChatbotTGChatView: Error getting chat: %v", err)
			return

		}
	} else {
		tmplVars.TGChat = chat
	}

	elapsed := time.Since(start)
	tmplVars.DebugTime = elapsed.String()
	tmplErr := tmpl.ExecuteTemplate(response, "layout", tmplVars)
	if tmplErr != nil {
		logger.Warningf("HandleChatbotTGChatView: template error: %s", err.Error())
	}

	elapsed = time.Since(start)
	logger.Tracef("HandleChatbotTGChatView() [%s]", elapsed)
	return
}
