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
	IsNewDevice bool
}

type UserInfo struct {
	*oidc.UserInfo
}

type Authenticator interface {
	Authenticate(context.Context) (*AuthResult, error)
	GetAuthData() *AuthData
	Revoke(context.Context) error
	GetUserInfo(context.Context) (*UserInfo, error)
	IsAuthenticationExpired() (bool, time.Time, time.Time)
}
