package web

import (
	"net/http"

	"../files"
	"../registry"
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
	// Init Session
	tmplVars := &TemplateVarFiles{}
	initSessionVars(response, request, tmplVars)

	tmplVars.IsInit = files.IsInit()
	tmpl, err := compileTemplates("templates/layout.html", "templates/files.html")
	if err != nil {
		MakeErrorResponse(response, 500, err.Error(), 0)
		return
	}

	tmpl.ExecuteTemplate(response, "layout", tmplVars)
}

// HandleFilesConfig displays files config page
func HandleFilesConfig(response http.ResponseWriter, request *http.Request) {
	// Init Session
	tmplVars := &TemplateVarFilesConfig{}
	us := initSessionVars(response, request, tmplVars)

	if request.Method == "POST" {
		err := request.ParseForm()
		if err != nil {
			logger.Errorf("Error parseing form: %v", err)
			return
		}

		formEndpoint := ""
		if val, ok := request.Form["endpoint"]; ok {
			formEndpoint = val[0]
		}
		formBucket := ""
		if val, ok := request.Form["bucket"]; ok {
			formBucket = val[0]
		}
		formKeyID := ""
		if val, ok := request.Form["key_id"]; ok {
			formKeyID = val[0]
		}
		formAccessKey := ""
		if val, ok := request.Form["access_key"]; ok {
			formAccessKey = val[0]
		}

		uid := us.Values["LoggedInUserID"].(int)

		// Get Parent or Create
		var regParent *registry.Entry
		regParent, err = registry.Get("/system/files")
		if err != nil {
			logger.Errorf("Error getting /system/files: %s", err.Error())
			if err == registry.ErrDoesNotExist {
				logger.Infof("Could not get /system/files, creating")
				var regSystem *registry.Entry
				regSystem, err2 := registry.Get("/system")
				if err2 != nil {
					if err == registry.ErrDoesNotExist {
						logger.Infof("Could not get /system, creating")
						var regRoot *registry.Entry
						regRoot, err3 := registry.Get("/")
						if err3 != nil {
							logger.Errorf("Could not get root: %s", err3.Error())
							MakeErrorResponse(response, 500, err.Error(), 0)
							return
						}
						var errNew error
						regSystem, errNew = registry.New(regRoot.ID, "system", "", false, uid)
						if errNew != nil {
							logger.Errorf("Could not create /system/files: %s", errNew.Error())
							MakeErrorResponse(response, 500, err.Error(), 0)
							return
						}

					} else {
						logger.Errorf("Could not get /system: %s", err2.Error())
						MakeErrorResponse(response, 500, err.Error(), 0)
						return
					}
				}
				var errNew error
				regParent, errNew = registry.New(regSystem.ID, "files", "", false, uid)
				if errNew != nil {
					logger.Errorf("Could not create /system/files: %s", errNew.Error())
					MakeErrorResponse(response, 500, err.Error(), 0)
					return
				}
			} else {
				logger.Errorf("Could not get /system: %s", err.Error())
				MakeErrorResponse(response, 500, err.Error(), 0)
				return
			}
		}

		// Get Registry Entries or Create
		var regEndpoint *registry.Entry
		regEndpoint, err = registry.Get("/system/files/endpoint")
		if err != nil {
			if err == registry.ErrDoesNotExist {
				var errNew error
				regEndpoint, errNew = registry.New(regParent.ID, "endpoint", "", false, uid)
				if errNew != nil {
					logger.Errorf("Could not create /system/files/endpoint: %s", errNew.Error())
					MakeErrorResponse(response, 500, err.Error(), 0)
					return
				}
			}
		}

		var regBucket *registry.Entry
		regBucket, err = registry.Get("/system/files/bucket")
		if err != nil {
			if err == registry.ErrDoesNotExist {
				var errNew error
				regBucket, errNew = registry.New(regParent.ID, "bucket", "", false, uid)
				if errNew != nil {
					logger.Errorf("Could not create /system/files/bucket: %s", errNew.Error())
					MakeErrorResponse(response, 500, err.Error(), 0)
					return
				}
			}
		}

		var regKeyID *registry.Entry
		regKeyID, err = registry.Get("/system/files/key_id")
		if err != nil {
			if err == registry.ErrDoesNotExist {
				var errNew error
				regKeyID, errNew = registry.New(regParent.ID, "key_id", "", false, uid)
				if errNew != nil {
					logger.Errorf("Could not create /system/files/key_id: %s", errNew.Error())
					MakeErrorResponse(response, 500, err.Error(), 0)
					return
				}
			}
		}

		var regAccessKey *registry.Entry
		regAccessKey, err = registry.Get("/system/files/access_key")
		if err != nil {
			if err == registry.ErrDoesNotExist {
				var errNew error
				regAccessKey, errNew = registry.New(regParent.ID, "access_key", "", true, uid)
				if errNew != nil {
					logger.Errorf("Could not create /system/files/access_key: %s", errNew.Error())
					MakeErrorResponse(response, 500, err.Error(), 0)
					return
				}
			}
		}

		// Set Values
		err = regEndpoint.SetValue(formEndpoint)
		if err != nil {
			logger.Errorf("Could not set /system/files/endpoint: %s", err.Error())
			MakeErrorResponse(response, 500, err.Error(), 0)
			return
		}
		err = regBucket.SetValue(formBucket)
		if err != nil {
			logger.Errorf("Could not set /system/files/bucket: %s", err.Error())
			MakeErrorResponse(response, 500, err.Error(), 0)
			return
		}
		err = regKeyID.SetValue(formKeyID)
		if err != nil {
			logger.Errorf("Could not set /system/files/key_id: %s", err.Error())
			MakeErrorResponse(response, 500, err.Error(), 0)
			return
		}
		err = regAccessKey.SetValue(formAccessKey)
		if err != nil {
			logger.Errorf("Could not set /system/files/access_key: %s", err.Error())
			MakeErrorResponse(response, 500, err.Error(), 0)
			return
		}

		files.InitClient(false)
	}

	logger.Tracef("HandleFilesConfig: Retrieving registry items")
	// Get Registry Entries
	var regEndpoint *registry.Entry
	regEndpoint, err := registry.Get("/system/files/endpoint")
	if err == nil {
		value, err := regEndpoint.GetValue()
		if err == nil {
			tmplVars.S3Endpoint = value
		}
	}

	var regBucket *registry.Entry
	regBucket, err = registry.Get("/system/files/bucket")
	if err == nil {
		value, err := regBucket.GetValue()
		if err == nil {
			tmplVars.BucketName = value
		}
	}

	var regKeyID *registry.Entry
	regKeyID, err = registry.Get("/system/files/key_id")
	if err == nil {
		value, err := regKeyID.GetValue()
		if err == nil {
			tmplVars.KeyID = value
		}
	}

	var regAccessKey *registry.Entry
	regAccessKey, err = registry.Get("/system/files/access_key")
	if err == nil {
		value, err := regAccessKey.GetValue()
		if err == nil {
			tmplVars.AccessKey = value
		}
	}

	tmplVars.IsInit = files.IsInit()
	tmpl, err := compileTemplates("templates/layout.html", "templates/files_config.html")
	if err != nil {
		MakeErrorResponse(response, 500, err.Error(), 0)
		return
	}

	tmpl.ExecuteTemplate(response, "layout", tmplVars)
}
