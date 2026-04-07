package errors

type Code string

const (
CodeInvalidArgument Code = "INVALID_ARGUMENT"
CodeNotFound        Code = "NOT_FOUND"
CodeConflict        Code = "CONFLICT"
CodeUnauthorized    Code = "UNAUTHORIZED"
CodeInternal        Code = "INTERNAL"
)
