package auth

import (
	"flag"
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

// --- Flags ---

var (
	authMethod      = flag.String("auth-method", "none", "Authentication method (none, token, basic)")
	authTokens      = flag.String("auth-tokens", "", "Comma-separated list of valid tokens (for token auth)")
	authUsername    = flag.String("auth-username", "", "Username for basic auth")
	authPassword    = flag.String("auth-password", "", "Password for basic auth")
)

func InitAuthFromFlags() *AuthModule {
	switch *authMethod {
	case "token":
		if *authTokens == "" {
			logger.Error("No tokens provided for token-based authentication")
			return nil
		}
		tokens := strings.Split(*authTokens, ",")
		return NewAuthModule(NewTokenAuth(tokens))

	case "basic":
		if *authUsername == "" || *authPassword == "" {
			logger.Error("Username and password are required for basic authentication")
			return nil
		}
		return NewAuthModule(NewBasicAuth(*authUsername, *authPassword))

	case "none":
		return NewAuthModule(&NoAuth{})

	default:
		logger.Error("Invalid authentication method: %s", *authMethod)
		return nil
	}
}