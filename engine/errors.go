package engine

import "github.com/hanchen1103/iteration_engine/domain"

func invalidError(message string) error {
	return domain.NewError(domain.ErrorCodeInvalid, message)
}

func notFoundError(message string) error {
	return domain.NewError(domain.ErrorCodeNotFound, message)
}

func conflictError(message string) error {
	return domain.NewError(domain.ErrorCodeConflict, message)
}

func failedError(message string) error {
	return domain.NewError(domain.ErrorCodeFailed, message)
}

func forbiddenError(message string) error {
	return domain.NewError(domain.ErrorCodeForbidden, message)
}
