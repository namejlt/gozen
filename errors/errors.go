// Package errors provides a unified error interface and error-code registry
// for the gozen framework. It is the single source of truth for application-level
// errors, supporting code lookup, localization, and hot-reload of code definitions.
package errors

import (
	"fmt"
	"strconv"
	"sync"
)

// Code constants — reserved ranges:
//
//	  0 –  999  success / common
//	1000 – 1999  access / auth
//	2000 – 2099  mysql
//	2100 – 2199  mongodb
//	2200 – 2299  redis
//	2300 – 2399  elasticsearch
//	2400 – 2499  external API
//	2500 – 2599  application-object (ao)
//	2600 – 2699  gRPC
//	2700 – 2799  TMQ
//	2800 – 2899  MQ
const (
	CodeOK               = 0    // success
	CodeCommonOK         = 1001 // success (public)
	CodeAccessFail       = 1002 // access denied
	CodeServerBusy       = 1003 // server busy
	CodeParamsIncomplete = 1004 // incomplete parameters
	CodeUserNoLogin      = 1005 // user not logged in
	CodeUserNoLoginApp   = 1024 // app user not logged in
	CodeBusinessError    = 1010 // business error
)

// CodeFalseSet contains codes that should be treated as "not-success"
// when evaluating whether a request result is "true" or "false".
var CodeFalseSet = map[int]struct{}{
	CodeServerBusy: {},
}

// --------------------------------------------------------------------------
// Error interface
// --------------------------------------------------------------------------

// Error is the standard error type for the framework.
// It carries an integer code, a human-readable message, and optionally
// wraps an underlying error.
type Error interface {
	error

	// Code returns the numeric error code.
	Code() int

	// Message returns the human-readable message.
	Message() string

	// Unwrap returns the wrapped error, if any.
	Unwrap() error
}

// AppError is the default implementation of Error.
type AppError struct {
	code int
	msg  string
	err  error
}

// New creates an AppError with the given code and optional cause.
// If msg is empty it will be looked up from the registry.
func New(code int, msg string, cause error) *AppError {
	if msg == "" {
		msg = GetMessage(code)
	}
	return &AppError{code: code, msg: msg, err: cause}
}

// Newf creates an AppError with a formatted message.
func Newf(code int, format string, args ...any) *AppError {
	return &AppError{code: code, msg: fmt.Sprintf(format, args...)}
}

func (e *AppError) Error() string {
	if e.err != nil {
		return fmt.Sprintf("[%d] %s: %v", e.code, e.msg, e.err)
	}
	return fmt.Sprintf("[%d] %s", e.code, e.msg)
}

func (e *AppError) Code() int        { return e.code }
func (e *AppError) Message() string  { return e.msg }
func (e *AppError) Unwrap() error    { return e.err }

// --------------------------------------------------------------------------
// Code registry
// --------------------------------------------------------------------------

var (
	codeMessages sync.Map
)

func init() {
	// Seed built-in codes.
	SetMessage(CodeOK, "success")
	SetMessage(CodeCommonOK, "success")
	SetMessage(CodeAccessFail, "access denied")
	SetMessage(CodeServerBusy, "server busy")
	SetMessage(CodeParamsIncomplete, "incomplete parameters")
	SetMessage(CodeUserNoLogin, "user not logged in")
	SetMessage(CodeUserNoLoginApp, "app user not logged in")
	SetMessage(CodeBusinessError, "business error")
}

// SetMessage stores a code → message mapping.
func SetMessage(code int, msg string) {
	codeMessages.Store(strconv.Itoa(code), msg)
}

// SetMessages bulk-loads code → message mappings (e.g. from config).
func SetMessages(m map[int]string) {
	for k, v := range m {
		SetMessage(k, v)
	}
}

// ResetMessages clears all code messages (useful for testing).
func ResetMessages() {
	codeMessages = sync.Map{}
}

// GetMessage returns the message for code, or "system error" if none found.
func GetMessage(code int) string {
	v, ok := codeMessages.Load(strconv.Itoa(code))
	if !ok {
		return "system error"
	}
	return v.(string)
}
