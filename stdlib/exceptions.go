package stdlib

import "fmt"

// Exception is the base exception type for Kylix exception handling.
// It implements the error interface for use with panic/recover.
type Exception struct {
	Message string
}

// Error implements the error interface.
func (e *Exception) Error() string {
	return e.Message
}

// NewException creates a new Exception with the given message.
func NewException(msg string) *Exception {
	return &Exception{Message: msg}
}

// EConvertError is raised when type conversion fails.
type EConvertError struct {
	Exception
}

// NewConvertError creates a new EConvertError.
func NewConvertError(msg string) *EConvertError {
	return &EConvertError{Exception: Exception{Message: msg}}
}

// ERangeError is raised when a value is out of range.
type ERangeError struct {
	Exception
}

// NewRangeError creates a new ERangeError.
func NewRangeError(msg string) *ERangeError {
	return &ERangeError{Exception: Exception{Message: msg}}
}

// EFileNotFound is raised when a file is not found.
type EFileNotFound struct {
	Exception
}

// NewFileNotFound creates a new EFileNotFound.
func NewFileNotFound(msg string) *EFileNotFound {
	return &EFileNotFound{Exception: Exception{Message: msg}}
}

// EArgumentException is raised when an argument is invalid.
type EArgumentException struct {
	Exception
}

// NewArgumentException creates a new EArgumentException.
func NewArgumentException(msg string) *EArgumentException {
	return &EArgumentException{Exception: Exception{Message: msg}}
}

func init() {
	// Register exception factory to prevent "imported and not used" error
	_ = fmt.Sprintf("%v", NewException(""))
}
