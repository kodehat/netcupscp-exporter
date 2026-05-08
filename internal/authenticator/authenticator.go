package authenticator

import (
	"context"
	"net/http"
)

type AuthResult struct {
	IsNewDevice  bool
	RefreshToken string
}

type Authenticator interface {
	Authenticate(context.Context) (*AuthResult, error)
	GetAuthenticatedClient() *http.Client
}
