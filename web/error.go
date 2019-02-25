package web

import (
	"net/http"
	"strconv"
)

// TemplateVarError holds template variables for MakeErrorResponse
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

// HandleNotFound displays a 404 page
func HandleNotFound(response http.ResponseWriter, request *http.Request) {
	MakeErrorResponse(response, http.StatusNotFound, request.URL.Path, 0)
	return
}

// MakeErrorResponse creates and displays an error page
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
