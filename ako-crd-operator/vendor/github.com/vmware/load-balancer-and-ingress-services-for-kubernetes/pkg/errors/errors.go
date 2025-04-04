package errors

type AkoError struct {
	msg string
}

func (a *AkoError) Error() string {
	return a.msg
}

func NewAkoError(msg string) error {
	return &AkoError{msg: msg}
}
