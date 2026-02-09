package errors

import (
	"errors"
	"fmt"
	"time"
)

// Sentinel errors
var (
	ErrCapabilityNotFound      = errors.New("capability not found")
	ErrProviderNotFound        = errors.New("provider not found")
	ErrCapabilityAlreadyExists = errors.New("capability already exists")
	ErrProviderAlreadyExists   = errors.New("provider already exists")
	ErrInsufficientPermission  = errors.New("insufficient permissions")
	ErrInvalidCapabilitySchema = errors.New("invalid capability schema")
	ErrInvalidProviderMetadata = errors.New("invalid provider metadata")
	ErrInvalidSignature        = errors.New("invalid signature")
	ErrIncompatibleVersion     = errors.New("incompatible version")
)

// ErrorCode represents standardized error codes
type ErrorCode string

const (
	ValidationError     ErrorCode = "VALIDATION_ERROR"
	Unauthorized        ErrorCode = "UNAUTHORIZED"
	Forbidden           ErrorCode = "FORBIDDEN"
	NotFound            ErrorCode = "NOT_FOUND"
	Conflict            ErrorCode = "CONFLICT"
	IncompatibleVersion ErrorCode = "INCOMPATIBLE_VERSION"
	InvalidSignature    ErrorCode = "INVALID_SIGNATURE"
	RateLimited         ErrorCode = "RATE_LIMITED"
	InternalError       ErrorCode = "INTERNAL_ERROR"
)

// ErrorDetail represents a detailed error for a specific field
type ErrorDetail struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// APIError represents a standardized API error response
type APIError struct {
	Error struct {
		Code    ErrorCode     `json:"code"`
		Message string        `json:"message"`
		Details []ErrorDetail `json:"details,omitempty"`
	} `json:"error"`
	Timestamp string `json:"timestamp"`
}

// NewAPIError creates a new APIError
func NewAPIError(code ErrorCode, message string, details ...ErrorDetail) *APIError {
	err := &APIError{
		Timestamp: time.Now().Format(time.RFC3339),
	}
	err.Error.Code = code
	err.Error.Message = message
	if len(details) > 0 {
		err.Error.Details = details
	}
	return err
}

// HTTPStatus returns the HTTP status code for the error
func (e *APIError) HTTPStatus() int {
	switch e.Error.Code {
	case ValidationError:
		return 400
	case Unauthorized:
		return 401
	case Forbidden:
		return 403
	case NotFound:
		return 404
	case Conflict:
		return 409
	case IncompatibleVersion:
		return 422
	case InvalidSignature:
		return 422
	case RateLimited:
		return 429
	case InternalError:
		return 500
	default:
		return 500
	}
}

// NewValidationError creates a validation error
func NewValidationError(message string, details ...ErrorDetail) *APIError {
	return NewAPIError(ValidationError, message, details...)
}

// NewUnauthorizedError creates an unauthorized error
func NewUnauthorizedError(message string) *APIError {
	if message == "" {
		message = "Missing or invalid authentication token"
	}
	return NewAPIError(Unauthorized, message)
}

// NewForbiddenError creates a forbidden error
func NewForbiddenError(message string) *APIError {
	if message == "" {
		message = "Insufficient permissions"
	}
	return NewAPIError(Forbidden, message)
}

// NewNotFoundError creates a not found error
func NewNotFoundError(resource string) *APIError {
	message := fmt.Sprintf("%s not found", resource)
	return NewAPIError(NotFound, message)
}

// NewConflictError creates a conflict error
func NewConflictError(message string) *APIError {
	return NewAPIError(Conflict, message)
}

// NewIncompatibleVersionError creates an incompatible version error
func NewIncompatibleVersionError(message string) *APIError {
	return NewAPIError(IncompatibleVersion, message)
}

// NewInvalidSignatureError creates an invalid signature error
func NewInvalidSignatureError() *APIError {
	return NewAPIError(InvalidSignature, "Registry snapshot signature is invalid")
}

// NewInternalError creates an internal server error
func NewInternalError(message string) *APIError {
	if message == "" {
		message = "Internal server error"
	}
	return NewAPIError(InternalError, message)
}
