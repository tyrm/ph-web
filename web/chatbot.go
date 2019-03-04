package web

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"../chatbot/telegram"
	"../models"
	"github.com/gorilla/mux"
)

// TemplateVarFiles holds template variables for HandleFiles
type TemplateVarChatbot struct {
	templateVarLayout

	IsInit bool
}
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

// TelegramIsInit returns true if telegram is connected
func (_ *TemplateVarChatbot) TelegramIsInit() bool {
	return telegram.IsInit()
}

// HandleChatbot displays files home
func HandleChatbot(response http.ResponseWriter, request *http.Request) {
	start := time.Now()

	// Init Session
	tmplVars := &TemplateVarChatbot{}
	tmpl, _ := initSessionVars(response, request, tmplVars, "templates/layout.html", "templates/chatbot.html")

	err := tmpl.ExecuteTemplate(response, "layout", tmplVars)
	if err != nil {
		logger.Warningf("HandleChatbot: template error: %s", err.Error())
	}

	elapsed := time.Since(start)
	logger.Tracef("HandleRegistryIndex() [%s]", elapsed)
	return
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
	userCount, err := models.GetUserCount()
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

	err = tmpl.ExecuteTemplate(response, "layout", tmplVars)
	if err != nil {
		logger.Warningf("HandleChatbot: template error: %s", err.Error())
	}

	elapsed := time.Since(start)
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

	tmplErr := tmpl.ExecuteTemplate(response, "layout", tmplVars)
	if tmplErr != nil {
		logger.Warningf("HandleChatbotTGChatView: template error: %s", err.Error())
	}

	elapsed := time.Since(start)
	logger.Tracef("HandleChatbotTGChatView() [%s]", elapsed)
	return
}

// HandleChatbotTGPhotoSizeView displays files home
func HandleChatbotTGPhotoSizeView(response http.ResponseWriter, request *http.Request) {
	start := time.Now()

	// Init Session
	tmplVars := &TemplateVarChatbot{}
	_, _ = initSessionVars(response, request, tmplVars)

	vars := mux.Vars(request)
	fileID, err := models.ReadTGPhotoSizeByFileID(vars["id"])
	if err != nil {
		if err == models.ErrDoesNotExist {
			MakeErrorResponse(response, 404, vars["id"], 0)
			return
		}
		MakeErrorResponse(response, 500, err.Error(), 0)
		logger.Errorf("HandleUserGet: Error getting PhotoSize: %v", err)
		return
	}

	body, err := telegram.GetFile(fileID)
	if err != nil {
		MakeErrorResponse(response, 500, err.Error(), 0)
		logger.Errorf("HandleUserGet: Error getting PhotoSize: %v", err)
		return
	}


	response.Write(body)

	//_ , _ = fmt.Fprintf(response, "Hello %v", fileID)

	elapsed := time.Since(start)
	logger.Tracef("HandleChatbotTGPhotoSizeView() [%s]", elapsed)
	return
}