package kv

import "errors"

var (
	ErrKeyNotFound   = errors.New("key not found")
	ErrInvalidKey    = errors.New("invalid key")
	ErrValueTooLarge = errors.New("value too large")
	ErrReadOnly      = errors.New("store is read-only")
	ErrClosed        = errors.New("store is closed")
	ErrInvalidTTL    = errors.New("invalid TTL value")
)
