package errors

import "fmt"

type AKOCRDOperatorError struct {
	HttpStatusCode int
	Reason         string
	Message        string
}

func (e AKOCRDOperatorError) Error() string {
	return fmt.Sprintf("Error code: %d: %s", e.HttpStatusCode, e.Message)
}
