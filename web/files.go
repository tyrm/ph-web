package web

import (
	"fmt"
	"net/http"

	"../files"
)

// TemplateVarFiles holds template variables for HandleFiles
type TemplateVarFiles struct {
	templateVarLayout

	IsInit bool
}

// TemplateVarFilesConfig holds template variables for HandleFilesConfig
type TemplateVarFilesConfig struct {
	templateVarLayout

	S3Endpoint string
	BucketName string
	KeyID      string
	AccessKey  string

	IsInit bool
}

// HandleFiles displays files home
func HandleFiles(response http.ResponseWriter, request *http.Request) {
	defer stsd.NewTiming().Send(fmt.Sprintf("%s.web.%s.HandleFiles", stsdPrefix, request.Method))
	// Init Session
	tmplVars := &TemplateVarFiles{}
	tmpl, _ := initSessionVars(response, request, tmplVars, "templates/layout.html", "templates/files.html")

	tmplVars.IsInit = files.IsInit()
	err := tmpl.ExecuteTemplate(response, "layout", tmplVars)
	if err != nil {
		logger.Warningf("HandleFiles: template error: %s", err.Error())
	}
}
