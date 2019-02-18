package web

import (
	"net/http"
	"time"
)

func HandleLogout(response http.ResponseWriter, request *http.Request) {
	start := time.Now()

	// Init Session
	us, err := globalSessions.Get(request, "session-key")
	if err != nil {
		MakeErrorResponse(response, 500, err.Error(), 0)
		return
	}

	// Delete session.
	us.Options.MaxAge = -1
	if err = us.Save(request, response); err != nil {
		logger.Errorf("Error saving session: %v", err)
		MakeErrorResponse(response, 500, err.Error(), 0)
		return
	}

	response.Header().Set("Location", "/login")
	response.WriteHeader(http.StatusFound)

	elapsed := time.Since(start)
	logger.Tracef("HandleLogout() [%s]", elapsed)
	return
}