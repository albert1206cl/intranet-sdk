// Package utils provides utility functions for the MiniEye Intranet SDK.
package utils

import (
	"fmt"
)

// ErrorCode represents an error code in the MiniEye Intranet SDK.
type ErrorCode int

// Define error codes
const (
	ErrCodeUnknown ErrorCode = iota
	ErrCodeInvalidInput
	ErrCodeUnauthorized
	ErrCodeForbidden
	ErrCodeNotFound
	ErrCodeAPIError
	ErrCodeNetworkError
	ErrCodeInternalError
)

// String returns the string representation of the error code.
func (c ErrorCode) String() string {
	switch c {
	case ErrCodeUnknown:
		return "unknown error"
	case ErrCodeInvalidInput:
		return "invalid input"
	case ErrCodeUnauthorized:
		return "unauthorized"
	case ErrCodeForbidden:
		return "forbidden"
	case ErrCodeNotFound:
		return "not found"
	case ErrCodeAPIError:
		return "API error"
	case ErrCodeNetworkError:
		return "network error"
	case ErrCodeInternalError:
		return "internal error"
	default:
		return "unknown error code"
	}
}

// SDKError represents an error returned by the MiniEye Intranet SDK.
type SDKError struct {
	Code    ErrorCode
	Message string
	Err     error
	// ServerCode 保留服务端返回的业务 code（如 gRPC code 数字），
	// 仅供程序化读取排查，不参与 Error() 输出；未设置时为 0。
	ServerCode int
}

// Error implements the error interface.
func (e *SDKError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code.String(), e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code.String(), e.Message)
}

// Unwrap returns the underlying error.
func (e *SDKError) Unwrap() error {
	return e.Err
}

// NewSDKError creates a new SDK error.
func NewSDKError(code ErrorCode, message string, err error) *SDKError {
	return &SDKError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// NewInvalidInputError creates a new invalid input error.
func NewInvalidInputError(message string, err error) *SDKError {
	return NewSDKError(ErrCodeInvalidInput, message, err)
}

// NewUnauthorizedError creates a new unauthorized error.
func NewUnauthorizedError(message string, err error) *SDKError {
	return NewSDKError(ErrCodeUnauthorized, message, err)
}

// NewForbiddenError creates a new forbidden error.
func NewForbiddenError(message string, err error) *SDKError {
	return NewSDKError(ErrCodeForbidden, message, err)
}

// NewNotFoundError creates a new not found error.
func NewNotFoundError(message string, err error) *SDKError {
	return NewSDKError(ErrCodeNotFound, message, err)
}

// NewAPIError creates a new API error.
func NewAPIError(message string, err error) *SDKError {
	return NewSDKError(ErrCodeAPIError, message, err)
}

// NewNetworkError creates a new network error.
func NewNetworkError(message string, err error) *SDKError {
	return NewSDKError(ErrCodeNetworkError, message, err)
}

// NewInternalError creates a new internal error.
func NewInternalError(message string, err error) *SDKError {
	return NewSDKError(ErrCodeInternalError, message, err)
}

// NewLoginError creates a new login error.
func NewLoginError(message string, err error) *SDKError {
	return NewSDKError(ErrCodeUnauthorized, "登录失败: "+message, err)
}

// NewTokenError creates a new token error.
func NewTokenError(message string, err error) *SDKError {
	return NewSDKError(ErrCodeUnauthorized, "令牌错误: "+message, err)
}

// NewValidationError creates a new validation error.
func NewValidationError(field string, message string) *SDKError {
	return NewSDKError(ErrCodeInvalidInput, field+": "+message, nil)
}