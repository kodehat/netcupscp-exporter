package authenticator

import "context"

type AuthData struct {
	AccessToken  string
	RefreshToken string
}

type AuthResult struct {
	AuthData    *AuthData
	IsNewDevice bool
}

type Authenticator interface {
	Authenticate(context.Context) (*AuthResult, error)
	Revoke(context.Context) error
}
