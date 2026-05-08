package authenticator

import (
	"context"
	"net/http"
)

type AuthData struct {
	AccessToken  string
	RefreshToken string
	Client       *http.Client
}

type AuthResult struct {
	IsNewDevice bool
}

type Authenticator interface {
	Authenticate(context.Context) (*AuthResult, error)
	GetAuthData() *AuthData
}
