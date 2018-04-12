package errors

type ModelError struct {
	Message string // description of the error
}

func (e *ModelError) Error() string {
	return e.Message
}

// IsModelError ..
func IsModelError(err interface{}) (bool) {
	_, ok := err.(*ModelError)
	return ok
}
