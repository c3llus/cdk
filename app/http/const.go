package http

import "time"

// Timeout
const (
	readTimeout  = time.Duration(2 * time.Second)
	writeTimeout = time.Duration(2 * time.Second)
)
