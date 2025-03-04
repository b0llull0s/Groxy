package auth

import (
	"net/http"
	"strings"

	"Groxy/logger"
)

type AuthModule struct {
	method AuthMethod
}

func NewAuthModule(method AuthMethod) *AuthModule {
	return &AuthModule{
		method: method,
	}
}

func (a *AuthModule) Authenticate(req *http.Request) bool {
	if a.method == nil {
		logger.Warning("No authentication method configured, allowing request")
		return true
	}

	authorized := a.method.Authenticate(req)
	if !authorized {
		logger.Warning("Request unauthorized: %s %s", req.Method, req.URL.String())
	}

	return authorized
}

// --- Authentication Methods ---

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