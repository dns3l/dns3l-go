package client

import "errors"

var (
	// ErrBadRequest indicates that the server responded with HTTP 400 Bad Request.
	ErrBadRequest = errors.New("bad request")
	// ErrNotFound indicates that the server responded with HTTP 404 Not Found.
	ErrNotFound = errors.New("not found")
	// ErrConflict indicates that the server responded with HTTP 409 Conflict.
	ErrConflict = errors.New("conflict")
)
