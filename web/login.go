package web

import (
	"html/template"
	"net/http"
)

type TemplateVarLogin struct {
	Error     string
	Username  string
}

func HandleLogin(response http.ResponseWriter, request *http.Request) {
	us, err := globalSessions.Get(request, "session-key")
	if err != nil {
		MakeErrorResponse(response, 500, err.Error(), 0)
		return
	}

	if us.Values["LoggedInUserID"] != nil {
		response.Header().Set("Location", "/")
		response.WriteHeader(http.StatusFound)
		return
	}

	if err = us.Save(request, response); err != nil {
		logger.Errorf("Error saving session: %v", err)
		MakeErrorResponse(response, 500, err.Error(), 0)
		return
	}

	tmlpStr, err := templates.FindString("templates/login.html")
	tmpl := template.New("login template")
	tmpl = template.Must(tmpl.Parse(tmlpStr))

	tmpl.Execute(response, &TemplateVarLogin{
		Error: "whoop",
	})
	return
}