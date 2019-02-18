package registry

const (
	LogAdded    = 1
	LogModified = 2
)

var logText = map[int]string{
	LogAdded:    "%s added '%s' with value '%s'.",
	LogModified: "%s updated '%s' from '%s' to '%s'.",
}

func getLogText(code int) string {
	return logText[code]
}