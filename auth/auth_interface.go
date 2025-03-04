package auth

import (
	"net/http"
	"strings"
)

type AuthMethod interface {
	Authenticate(req *http.Request) bool
}

type NoAuth struct{}

func (n *NoAuth) Authenticate(req *http.Request) bool {
	return true
}

// TokenAuth 
type TokenAuth struct {
	ValidTokens map[string]bool
}

func NewTokenAuth(validTokens []string) *TokenAuth {
	tokens := make(map[string]bool)
	for _, token := range validTokens {
		tokens[token] = true
	}
	return &TokenAuth{
		ValidTokens: tokens,
	}
}

func (t *TokenAuth) Authenticate(req *http.Request) bool {
	token := req.Header.Get("Authorization")
	if token == "" {
		return false
	}

	token = strings.TrimPrefix(token, "Bearer ")

	_, valid := t.ValidTokens[token]
	return valid
}

// BasicAuth
type BasicAuth struct {
	Username string
	Password string
}

func NewBasicAuth(username, password string) *BasicAuth {
	return &BasicAuth{
		Username: username,
		Password: password,
	}
}

func (b *BasicAuth) Authenticate(req *http.Request) bool {
	username, password, ok := req.BasicAuth()
	if !ok {
		return false
	}

	return username == b.Username && password == b.Password
}