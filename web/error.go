package web

import (
	"net/http"
	"strconv"
)

type TemplateVarError struct {
	ErrNum  string
	CodeNum string
	ErrText string
	Detail  string
}

var codeTitle = map[int]string{
	1:    "Malformed JSON Body",
	2201: "Missing Required Attribute",
	2202: "Requested Relationship Not Found",
}

func HandleNotFound(response http.ResponseWriter, request *http.Request) {
	MakeErrorResponse(response, http.StatusNotFound, request.URL.Path, 0)
	return
}

func MakeErrorResponse(response http.ResponseWriter, status int, detail string, code int) {
	templateVars := &TemplateVarError{
		ErrNum: strconv.Itoa(status),
		Detail: detail,
	}

	// Get Title
	if code == 0 { // code 0 means no code
		templateVars.ErrText = http.StatusText(status)
	} else {
		templateVars.ErrText = codeTitle[code]
		templateVars.CodeNum = strconv.Itoa(code)
	}

	// Send Response
	response.WriteHeader(status)

	tmpl, err := compileTemplates("templates/error.html")
	if err != nil {
		MakeErrorResponse(response, 500, err.Error(), 0)
		return
	}

	tmpl.Execute(response, templateVars)

	return
}

func ProtectMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		us, err := globalSessions.Get(r, "session-key")
		if err != nil {
			MakeErrorResponse(w, 500, err.Error(), 0)
			return
		}

		if us.Values["LoggedInUserID"] == nil {
			MakeErrorResponse(w, 404, r.URL.Path, 0)
			return
		}

		next.ServeHTTP(w, r)
	})
}
