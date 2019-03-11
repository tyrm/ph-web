package web

import (
	"fmt"
	"html"
	"net/http"
	"strconv"
	"time"

	"../models"
	"github.com/gorilla/mux"
)

// TemplateVarUserIndex holds template variables for HandleUserIndex
type TemplateVarUserIndex struct {
	templateVarLayout

	Users []*models.User
	Pages *templatePages
}

// TemplateVarUserNew holds template variables for HandleUserNew
type TemplateVarUserNew struct {
	templateVarLayout
}

// TemplateVarUserView holds template variables for HandleUserGet
type TemplateVarUserView struct {
	User *models.User
	templateVarLayout
}

// HandleUserGet displays information about a user
func HandleUserGet(response http.ResponseWriter, request *http.Request) {
	start := time.Now()

	// Init Session
	tmplVars := &TemplateVarUserView{}
	tmpl, _ := initSessionVars(response, request, tmplVars, "templates/layout.html", "templates/users_get.html")

	vars := mux.Vars(request)
	user, err := models.ReadUser(vars["id"])
	if err != nil {
		MakeErrorResponse(response, 500, err.Error(), 0)
		logger.Errorf("HandleUserGet: Error getting user: %v", err)
		return
	}

	tmplVars.User = user

	elapsed := time.Since(start)
	tmplVars.DebugTime = elapsed.String()
	err = tmpl.ExecuteTemplate(response, "layout", tmplVars)
	if err != nil {
		logger.Warningf("HandleUserGet: template error: %s", err.Error())
	}

	elapsed = time.Since(start)
	logger.Tracef("HandleUserGet() [%s]", elapsed)
	return
}

// HandleUserIndex displays a list of users
func HandleUserIndex(response http.ResponseWriter, request *http.Request) {
	start := time.Now()

	// Init Session
	tmplVars := &TemplateVarUserIndex{}
	tmpl, us := initSessionVars(response, request, tmplVars, "templates/layout.html", "templates/users_index.html")


	if request.Method == "POST" {
		err := request.ParseForm()
		if err != nil {
			tmplVars.AlertError = fmt.Sprintf("error parsing form: %s", html.EscapeString(err.Error()))
		} else {
			logger.Tracef("got post: %v", request.Form)

			if val, ok := request.Form["_action"]; ok {
				formAction := val[0]

				if formAction == "delete" {
					formUserToken := ""
					if val, ok := request.Form["user_token"]; ok {
						formUserToken = val[0]
					}

					if formUserToken != "" {
						err := models.DeleteUser(formUserToken)
						if err != nil {
							tmplVars.AlertError = fmt.Sprintf("error deleting user: %s", html.EscapeString(err.Error()))
						} else {
							tmplVars.AlertSuccess = fmt.Sprintf("user successfully deleted.")
						}
					}
				} else {
					tmplVars.AlertError = fmt.Sprintf("unknown action: %s", html.EscapeString(formAction))
				}
			} else {
				tmplVars.AlertError = fmt.Sprintf("missing action")
				return
			}

		}
	}

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
	users, err := models.ReadUsersPage(entriesPerPage, page-1)
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

	elapsed := time.Since(start)
	tmplVars.DebugTime = elapsed.String()
	err = tmpl.ExecuteTemplate(response, "layout", tmplVars)
	if err != nil {
		logger.Warningf("HandleUserIndex: template error: %s", err.Error())
	}

	elapsed = time.Since(start)
	logger.Tracef("HandleUserIndex() [%s]", elapsed)
	return
}

// HandleUserNew handles creating a new user
func HandleUserNew(response http.ResponseWriter, request *http.Request) {
	start := time.Now()

	// Init Session
	tmplVars := &TemplateVarUserIndex{}
	tmpl, _ := initSessionVars(response, request, tmplVars, "templates/layout.html", "templates/users_new.html")

	if request.Method == "POST" {
		err := request.ParseForm()
		if err != nil {
			logger.Errorf("Error parseing form: %v", err)
			return
		}

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
			newUser, err := models.CreateUser(formUsername, formPassword1, formEmail)
			if err != nil {
				tmplVars.AlertError = err.Error()
			} else {
				newPage := fmt.Sprintf("/web/admin/users/%s", newUser.Token)

				response.Header().Set("Location", newPage)
				response.WriteHeader(http.StatusFound)
				return
			}
		}
	}

	elapsed := time.Since(start)
	tmplVars.DebugTime = elapsed.String()
	err := tmpl.ExecuteTemplate(response, "layout", tmplVars)
	if err != nil {
		logger.Warningf("HandleUserNew: template error: %s", err.Error())
	}

	elapsed = time.Since(start)
	logger.Tracef("HandleUserNew() [%s]", elapsed)
	return
}
