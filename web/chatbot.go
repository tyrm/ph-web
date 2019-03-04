package web

import (
	"net/http"
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

// HandleChatbotTGPhotoSizeViewFile displays files home
func HandleChatbotTGPhotoSizeViewFile(response http.ResponseWriter, request *http.Request) {
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

	body, err := telegram.GetPhotoFile(fileID)
	if err != nil {
		MakeErrorResponse(response, 500, err.Error(), 0)
		logger.Errorf("HandleUserGet: Error getting PhotoSize: %v", err)
		return
	}


	response.Write(body)

	//_ , _ = fmt.Fprintf(response, "Hello %v", fileID)

	elapsed := time.Since(start)
	logger.Tracef("HandleChatbotTGPhotoSizeViewFile() [%s]", elapsed)
	return
}


// HandleChatbotTGPhotoSizeViewFile displays files home
func HandleChatbotTGStickerViewFile(response http.ResponseWriter, request *http.Request) {
	start := time.Now()

	vars := mux.Vars(request)
	sticker, err := models.ReadTGStickerByFileID(vars["id"])
	if err != nil {
		if err == models.ErrDoesNotExist {
			MakeErrorResponse(response, 404, vars["id"], 0)
			return
		}
		MakeErrorResponse(response, 500, err.Error(), 0)
		logger.Errorf("HandleUserGet: Error getting Stioker: %v", err)
		return
	}

	body, err := telegram.GetStickerFile(sticker)
	if err != nil {
		MakeErrorResponse(response, 500, err.Error(), 0)
		logger.Errorf("HandleUserGet: Error getting Stioker: %v", err)
		return
	}

	response.Write(body)

	elapsed := time.Since(start)
	logger.Tracef("HandleChatbotTGPhotoSizeViewFile() [%s]", elapsed)
	return
}

