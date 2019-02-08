package web

import (
	"net/http"
)

func HandleLogout(response http.ResponseWriter, request *http.Request) {
	us, err := globalSessions.Get(request, "session-key")

	// Delete session.
	us.Options.MaxAge = -1
	if err = us.Save(request, response); err != nil {
		logger.Errorf("Error saving session: %v", err)
		MakeErrorResponse(response, 500, err.Error(), 0)
		return
	}

	response.Header().Set("Location", "/login")
	response.WriteHeader(http.StatusFound)
	return
}