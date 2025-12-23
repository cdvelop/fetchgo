package fetch

var (
	logger        func(...any)
	globalHandler func(*Response)
)

// SetLog sets the logger function for debugging.
func SetLog(fn func(...any)) {
	logger = fn
}

// SetHandler sets the global handler for Dispatch requests.
func SetHandler(fn func(*Response)) {
	globalHandler = fn
}

// log prints a message if a logger is set.
func log(args ...any) {
	if logger != nil {
		logger(args...)
	}
}
