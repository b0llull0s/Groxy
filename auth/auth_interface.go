package auth

import (
	"net/http"
)

type AuthMethod interface {
	Authenticate(req *http.Request) bool
}

// Dummy method
type NoAuth struct{}

func (n *NoAuth) Authenticate(req *http.Request) bool {
	return true
}