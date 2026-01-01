package authenticator

import (
	"context"
	"time"

	"github.com/zitadel/oidc/v3/pkg/oidc"
)

type AuthData struct {
	AccessToken  string
	RefreshToken string
	TokenType    string
	Expiry       time.Time
	Subject      string
}

type AuthResult struct {
	AuthData    *AuthData
	IsNewDevice bool
}

type UserInfo struct {
	*oidc.UserInfo
}

type Authenticator interface {
	Authenticate(context.Context) (*AuthResult, error)
	Revoke(context.Context) error
	GetUserInfo(context.Context) (*UserInfo, error)
}
