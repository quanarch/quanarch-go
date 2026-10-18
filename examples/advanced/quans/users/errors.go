package users

// appError is a failure the quan expects and answers on purpose, with its own status and code.
type appError struct {
	status        int
	code, message string
}

func (e *appError) Error() string { return e.message }
