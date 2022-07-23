package middleware

import "encoding/json"

var ErrNotFound = NewAppError(nil, "not found")

type AppError struct {
	Err     error  `json:"_,omitempty"`
	Message string `json:"message,omitempty"`
	Code    string `json:"code,omitempty"`
}

func (e *AppError) Error() string {
	return e.Message
}

func (e *AppError) Marshal() []byte {
	marshal, err := json.Marshal(e)
	if err != nil {
		return nil
	}
	return marshal
}

func NewAppError(err error, message string) *AppError {
	return &AppError{
		Err:     err,
		Message: message,
	}
}

func systemError(err error) *AppError {
	return NewAppError(err, "internal system error")

}
