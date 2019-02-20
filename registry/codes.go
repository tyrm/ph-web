package registry

const (
	LogAdded          = 1
	LogModified       = 2
	LogAddedSecure    = 3
	LogModifiedSecure = 4
)

var logText = map[int]string{
	LogAdded:    "%s added '%s' with value '%s'.",
	LogModified: "%s updated '%s' from '%s' to '%s'.",
	LogAddedSecure: "%s added secure entry '%s'.",
	LogModifiedSecure: "$s updated secure entry '%s'.",
}

func getLogText(code int) string {
	return logText[code]
}

func logChange(rid int, uid int, changeType int, oldValue string, newValue string) {
	err := db.QueryRow(sqlCreateLog, rid, uid, changeType, oldValue, newValue).Scan()
	if err != nil {
		logger.Errorf("logChange: %s", err.Error())
	}
}
