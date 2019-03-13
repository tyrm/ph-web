package web


import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/bamzi/jobrunner"
)

// TemplateVarHome holds template variables for HandleHome
type TemplateVarJobrunner struct {
	templateVarLayout

	Jobs *[]jobrunner.StatusData
}
type TemplateVarJobrunnerEntry struct {
	Name string

}

// HandleHome displays the home dashboard
func HandleJobrunner(response http.ResponseWriter, request *http.Request) {
	defer stsd.NewTiming().Send(fmt.Sprintf("%s.web.%s.HandleJobrunner", stsdPrefix, request.Method))
	// Init Session
	tmplVars := &TemplateVarJobrunner{}
	tmpl, _ := initSessionVars(response, request, tmplVars, "templates/layout.html", "templates/jobrunner.html")

	entries := jobrunner.StatusPage()
	tmplVars.Jobs = &entries

	tmplVars.MetaRefresh = 10

	err := tmpl.ExecuteTemplate(response, "layout", tmplVars)
	if err != nil {
		logger.Warningf("HandleJobrunner: template error: %s", err.Error())
	}
	return
}

// HandleHome displays the home dashboard
func HandleJobrunnerJSON(response http.ResponseWriter, request *http.Request) {
	defer stsd.NewTiming().Send(fmt.Sprintf("%s.web.%s.HandleJobrunnerJSON", stsdPrefix, request.Method))

	err := json.NewEncoder(response).Encode(jobrunner.StatusJson())
	if err != nil {
		logger.Warningf("HandleJobrunnerJSON: template error: %s", err.Error())
	}

	return
}