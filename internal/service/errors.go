package service

import "errors"

// Service domain errors
var (
	ErrNotFound            = errors.New("not found")
	ErrRecordAlreadyExists = errors.New("record already exists")
	ErrURLAlreadyExists    = errors.New("URL already exists")
	ErrURLDeleterStopped   = errors.New("URL deleter stopped")
	ErrDeleted             = errors.New("URL deleted")
)
