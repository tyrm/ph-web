package web

import (
	"database/sql"
	"net/http"
	"time"

	"../models"
)

type TemplateVarLogin struct {
	Error     string
	Username  string
}

func HandleLogin(response http.ResponseWriter, request *http.Request) {
	start := time.Now()

	// Init Session
	us, err := globalSessions.Get(request, "session-key")
	if err != nil {
		MakeErrorResponse(response, 500, err.Error(), 0)
		return
	}

	tmpl, err := compileTemplates("templates/login.html")
	if err != nil {
		MakeErrorResponse(response, 500, err.Error(), 0)
		return
	}

	if request.Method == "POST" {
		request.ParseForm()
		formUsername := request.Form["username"][0]
		logger.Tracef("Trying login for: %s", formUsername)

		user, err := models.ReadUserByUsername(formUsername)
		if err == sql.ErrNoRows {
			tmpl.Execute(response, &TemplateVarLogin{Error: "username/password not recognized", Username: formUsername})
			return
		} else if err != nil {
			logger.Errorf("Couldn't get user for login: %s", err)
			tmpl.Execute(response, &TemplateVarLogin{Error: err.Error(), Username: formUsername})
			return
		}

		formPassword := request.Form["password"][0]
		valid := user.CheckPassword(formPassword)

		if valid {
			user.UpdateLastLogin()
			us.Values["LoggedInUserID"] = user.ID
		} else {
			tmpl.Execute(response, &TemplateVarLogin{Error: "username/password not recognized", Username: formUsername})
			return
		}
	}

	if err = us.Save(request, response); err != nil {
		logger.Errorf("Error saving session: %v", err)
		MakeErrorResponse(response, 500, err.Error(), 0)
		return
	}

	if us.Values["LoggedInUserID"] != nil {
		response.Header().Set("Location", "/web/")
		response.WriteHeader(http.StatusFound)
		return
	}

	err = tmpl.Execute(response, &TemplateVarLogin{})
	if err != nil {
		logger.Errorf("HandleLogin: Error executing template: %v", err)
	}

	elapsed := time.Since(start)
	logger.Tracef("HandleLogin() [%s]", elapsed)
	return
}