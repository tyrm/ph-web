package web

import (
	"fmt"
	"net/http"
)

// ProtectMiddleware redirects users who aren't logged in to the login page
func ProtectMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer stsd.NewTiming().Send(fmt.Sprintf("%s.web.ProtectMiddleware", stsdPrefix))
		us, err := globalSessions.Get(r, "session-key")
		if err != nil {
			MakeErrorResponse(w, 500, err.Error(), 0)
			return
		}

		if us.Values["LoggedInUserID"] == nil {
			if r.URL.RawQuery != "" {
				logger.Tracef("Path [%s?%s]", r.URL.Path, r.URL.RawQuery)
				us.Values["LastURL"] = fmt.Sprintf("%s?%s", r.URL.Path, r.URL.RawQuery)
			} else {
				logger.Tracef("Path [%s]", r.URL.Path)
				us.Values["LastURL"] = fmt.Sprintf("%s", r.URL.Path)
			}

			if err = us.Save(r, w); err != nil {
				logger.Errorf("Error saving session: %v", err)
				MakeErrorResponse(w, 500, err.Error(), 0)
				return
			}

			w.Header().Set("Location", "/login")
			w.WriteHeader(http.StatusFound)
			return
		}

		if val, ok := r.URL.Query()["dark_mode"]; ok {
			darkMode := val[0]

			if darkMode == "true" {
				us.Values["TemplateDarkMode"] = true
			} else if darkMode == "false" {
				us.Values["TemplateDarkMode"] = false
			}

			if err = us.Save(r, w); err != nil {
				logger.Errorf("Error saving session: %v", err)
				MakeErrorResponse(w, 500, err.Error(), 0)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}