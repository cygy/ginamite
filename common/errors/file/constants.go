// Package file lists the error codes of the "file" domain.
package file

// Domain is the name of the domain of the errors.
const Domain = "file"

// Error codes of the domain.
const (
	// CanNotReadFile is the code of the error raised when the file can not be read.
	CanNotReadFile = 0

	// CanNotSaveFile is the code of the error raised when the file can not be saved.
	CanNotSaveFile = 1

	// FileNotFound is the code of the error raised when the file is not found.
	FileNotFound = 2

	// MaxFileSizeExceeded is the code of the error raised when the file is bigger than the maximum size.
	MaxFileSizeExceeded = 3

	// InvalidMimeType is the code of the error raised when the MIME type of the file is not allowed.
	InvalidMimeType = 4
)
