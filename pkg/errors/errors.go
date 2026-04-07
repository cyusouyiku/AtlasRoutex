package errors

import "fmt"

type AppError struct {
Code    Code
Message string
Cause   error
}

func (e *AppError) Error() string {
if e.Cause == nil {
return fmt.Sprintf("%s: %s", e.Code, e.Message)
}
return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Cause)
}

func (e *AppError) Unwrap() error { return e.Cause }

func New(code Code, msg string, cause error) *AppError {
return &AppError{Code: code, Message: msg, Cause: cause}
}
