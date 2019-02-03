package web

import (
	"html/template"
	"net/http"
	"strconv"
)

type TemplateVarError struct {
	ErrNum  string
	CodeNum string
	ErrText string
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

	tmpl := template.Must(template.ParseFiles("templates/error.html"))
	tmpl.Execute(response, templateVars)

	return
}