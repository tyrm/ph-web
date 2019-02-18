package web

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"../models"
	"github.com/gorilla/mux"
)

type TemplateVarUserIndex struct {
	TemplateVarLayout

	Users     []*models.User
	Pages     *TemplatePages
}

type TemplateVarUserNew struct {
	TemplateVarLayout
}

type TemplateVarUserView struct {
	User *models.User
	TemplateVarLayout
}

func HandleUserGet(response http.ResponseWriter, request *http.Request) {
	start := time.Now()

	// Init Session
	us, err := globalSessions.Get(request, "session-key")
	if err != nil {
		MakeErrorResponse(response, 500, err.Error(), 0)
		return
	}

	vars := mux.Vars(request)
	user, err := models.GetUser(vars["id"])
	if err != nil {
		MakeErrorResponse(response, 500, err.Error(), 0)
		logger.Errorf("HandleUserGet: Error getting user: %v", err)
		return
	}

	tmplVars := &TemplateVarUserView{
		User: user,
	}
	uid := us.Values["LoggedInUserID"].(string)
	tmplVars.Username = models.GetUsernameByID(uid)
	tmplVars.NavBar = makeNavbar(request.URL.Path)

	tmpl, err := compileTemplates("templates/layout.html", "templates/users_get.html")
	if err != nil {
		MakeErrorResponse(response, 500, err.Error(), 0)
		return
	}

	err = tmpl.ExecuteTemplate(response, "layout", tmplVars)
	if err != nil {
		logger.Errorf("HandleUserGet: Error executing template: %v", err)
	}

	elapsed := time.Since(start)
	logger.Tracef("HandleUserGet() [%s]", elapsed)
	return
}

func HandleUserIndex(response http.ResponseWriter, request *http.Request) {
	start := time.Now()

	// Init Session
	us, err := globalSessions.Get(request, "session-key")
	if err != nil {
		MakeErrorResponse(response, 500, err.Error(), 0)
		return
	}

	tmplVars := &TemplateVarUserIndex{}
	uid := us.Values["LoggedInUserID"].(string)
	tmplVars.Username = models.GetUsernameByID(uid)
	tmplVars.NavBar = makeNavbar(request.URL.Path)

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

	err = tmpl.ExecuteTemplate(response, "layout", tmplVars)
	if err != nil {
		logger.Errorf("HandleUserIndex: Error executing template: %v", err)
	}

	elapsed := time.Since(start)
	logger.Tracef("HandleUserIndex() [%s]", elapsed)
	return
}

func HandleUserNew(response http.ResponseWriter, request *http.Request) {
	start := time.Now()

	// Init Session
	us, err := globalSessions.Get(request, "session-key")
	if err != nil {
		MakeErrorResponse(response, 500, err.Error(), 0)
		return
	}

	tmplVars := &TemplateVarUserIndex{}
	uid := us.Values["LoggedInUserID"].(string)
	tmplVars.Username = models.GetUsernameByID(uid)
	tmplVars.NavBar = makeNavbar(request.URL.Path)

	if request.Method == "POST" {
		logger.Errorf("trying to parse form: %v", err)
		err = request.ParseForm()
		if err != nil {
			logger.Errorf("Error parseing form: %v", err)
			return
		}
		logger.Tracef("%v", request.Form)

		formUsername := ""
		if val, ok := request.Form["username"]; ok {
			formUsername = val[0]
		}
		formEmail := ""
		if val, ok := request.Form["email"]; ok {
			formEmail = val[0]
		}
		formPassword1 := ""
		if val, ok := request.Form["password1"]; ok {
			formPassword1 = val[0]
		}
		formPassword2 := ""
		if val, ok := request.Form["password2"]; ok {
			formPassword2 = val[0]
		}

		logger.Tracef("%s, [%s], [%s]", formUsername, formPassword1, formPassword2)

		usernameExists, err := models.GetUsernameExists(formUsername)
		if err != nil {
			logger.Errorf("Error chekcing for username: %v", err)
			MakeErrorResponse(response, 500, err.Error(), 0)
			return
		}
		if usernameExists {
			tmplVars.AlertError = "Username taken."
		} else if formUsername == "" {
			tmplVars.AlertError = "Username missing."
		} else if formEmail == "" {
			tmplVars.AlertError = "Email missing."
		} else if formPassword1 == "" || formPassword2 == "" {
			tmplVars.AlertError = "Password missing."
		} else if formPassword1 != formPassword2 {
			tmplVars.AlertError = "Passwords don't match."
		} else {
			newUser, err := models.NewUser(formUsername, formPassword1, formEmail)
			if err != nil {
				tmplVars.AlertError = err.Error()
			} else {
				newPage := fmt.Sprintf("/web/users/%s", newUser.Token)

				response.Header().Set("Location", newPage)
				response.WriteHeader(http.StatusFound)
				return

			}
		}

	}

	tmpl, err := compileTemplates("templates/layout.html", "templates/users_new.html")
	if err != nil {
		MakeErrorResponse(response, 500, err.Error(), 0)
		return
	}

	err = tmpl.ExecuteTemplate(response, "layout", tmplVars)
	if err != nil {
		logger.Errorf("HandleUserNew: Error executing template: %v", err)
	}

	elapsed := time.Since(start)
	logger.Tracef("HandleUserNew() [%s]", elapsed)
	return
}
