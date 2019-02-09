package web

import (
	"database/sql"
	"net/http"

	"../models"
)

type TemplateVarLogin struct {
	Error     string
	Username  string
}

func HandleLogin(response http.ResponseWriter, request *http.Request) {
	tmpl, err := compileTemplates("templates/login.html")
	if err != nil {
		MakeErrorResponse(response, 500, err.Error(), 0)
		return
	}

	us, err := globalSessions.Get(request, "session-key")
	if err != nil {
		MakeErrorResponse(response, 500, err.Error(), 0)
		return
	}

	if request.Method == "POST" {
		request.ParseForm()
		formUsername := request.Form["username"][0]
		logger.Tracef("Trying login for: %s", formUsername)

		user, err := models.GetUserByUsername(formUsername)
		if err == sql.ErrNoRows {
			tmpl.Execute(response, &TemplateVarLogin{Error: "username/password not recognized"})
			return
		} else if err != nil {
			logger.Errorf("Couldn't get user for login: %s", err)
			tmpl.Execute(response, &TemplateVarLogin{Error: err.Error()})
			return
		}

		formPassword := request.Form["password"][0]
		valid := user.CheckPassword(formPassword)

		if valid {
			us.Values["LoggedInUserID"] = user.ID
		} else {
			tmpl.Execute(response, &TemplateVarLogin{Error: "username/password not recognized"})
			return
		}
	}

	if err = us.Save(request, response); err != nil {
		logger.Errorf("Error saving session: %v", err)
		MakeErrorResponse(response, 500, err.Error(), 0)
		return
	}

	if us.Values["LoggedInUserID"] != nil {
		response.Header().Set("Location", "/")
		response.WriteHeader(http.StatusFound)
		return
	}

	tmpl.Execute(response, &TemplateVarLogin{})
	return
}