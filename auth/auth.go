package auth

import (
	"flag"
	"strings"
	"fmt"
	"net/http"
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
		fmt.Println("⚠️ WARNING: Running proxy with NO AUTHENTICATION. This is not recommended for production!")
		return NewAuthModule(&NoAuth{})

	default:
		logger.Error("Invalid authentication method: %s", *authMethod)
		return nil
	}
}


