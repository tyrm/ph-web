package web

import (
	"html/template"
	"net/http"
)

func HandleLanding(response http.ResponseWriter, request *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/landing.html"))
	tmpl.Execute(response, nil)
	return
}