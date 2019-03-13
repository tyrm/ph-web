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

type TemplateVarChatbotTGChatList struct {
	templateVarLayout

	Chats []*models.TGChat
	Pages *templatePages
}

type TemplateVarChatbotTGChatView struct {
	templateVarLayout

	TGChat *models.TGChat
	TGMessages []*TemplateVarChatbotMessageBlock

	IsInit bool
}

type TemplateVarChatbotMessageBlock struct {
	IsMe bool

	BlockMessages []*models.TGMessage
	BlockUser *models.TGUser

	ColorNum int
}

// HandleChatbot displays files home
func HandleChatbotTGChatList(response http.ResponseWriter, request *http.Request) {
	defer stsd.NewTiming().Send(fmt.Sprintf("%s.web.%s.HandleChatbotTGChatList", stsdPrefix, request.Method))
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
	defer stsd.NewTiming().Send(fmt.Sprintf("%s.web.%s.HandleChatbotTGChatView", stsdPrefix, request.Method))
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

		messages, err := chat.GetMessagesPage(20,0)
		if err == models.ErrDoesNotExist {
			tmplVars.AlertError = fmt.Sprintf("could not retrieve messages: %s", err)
		} else {
			if len(messages) > 0 {
				msgBlocks, err := makeMessageBlocks(messages)
				if err == models.ErrDoesNotExist {
					tmplVars.AlertError = fmt.Sprintf("could not build message blocks: %s", err)
				} else {
					tmplVars.TGMessages = msgBlocks
				}
			}
		}
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

func makeMessageBlocks(msgs []*models.TGMessage) ([]*TemplateVarChatbotMessageBlock, error) {
	defer stsd.NewTiming().Send(fmt.Sprintf("%s.web.makeMessageBlocks", stsdPrefix))

	var blockList []*TemplateVarChatbotMessageBlock
	var lastFrom int64 = -1

	var colorDB map[int]int
	colorDB = make(map[int]int)
	lastColor := 0

	var newBlock *TemplateVarChatbotMessageBlock

	for _, msg := range msgs {
		// check if we need a new block
		if !msg.FromID.Valid && lastFrom != 0  {
			if newBlock != nil {
				blockList = append(blockList, newBlock)
			}

			// create empty block
			newBlock = &TemplateVarChatbotMessageBlock{}

			lastFrom = 0
		} else if lastFrom != msg.FromID.Int64 {
			if newBlock != nil {
				blockList = append(blockList, newBlock)
			}

			// create new block with user
			fromUser, err := msg.GetFromUser()
			if err != nil {
				logger.Debugf("makeMessageBlocks(%d) (nil, %v)",len(msgs), err)
				return nil, err
			}

			fromUserPhoto, err := telegram.GetUserProfilePhotoCurrent(fromUser.APIID, 64)
			if err != nil {
				logger.Warningf("makeMessageBlocks: could not get photo url: %s", err)
			} else {
				fromUser.ProfilePhotoURL = fromUserPhoto
			}

			newColorNum := 0
			if val, ok := colorDB[fromUser.ID]; ok {
				newColorNum = val
			} else {
				newColorNum = lastColor + 1
				colorDB[fromUser.ID] = newColorNum
				if lastColor < 9 {
					lastColor = lastColor +1
				} else {
					lastColor = 0
				}
			}

			newBlock = &TemplateVarChatbotMessageBlock{
				BlockUser: fromUser,
				ColorNum: newColorNum,
			}

			lastFrom = msg.FromID.Int64
		}

		// add message to block
		newBlock.BlockMessages = append(newBlock.BlockMessages, msg)
	}

	blockList = append(blockList, newBlock)

	logger.Tracef("makeMessageBlocks(%d) (%d, nil)",len(msgs), len(blockList))
	return blockList, nil
}