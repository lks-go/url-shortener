package service

import "errors"

var (
	ErrNotFound            = errors.New("not found")
	ErrRecordAlreadyExists = errors.New("record already exists")
	ErrURLAlreadyExists    = errors.New("URL already exists")
)
