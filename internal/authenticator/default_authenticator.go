package authenticator

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/zitadel/oidc/v3/pkg/client/rp"
	"github.com/zitadel/oidc/v3/pkg/oidc"
)

var (
	netcupIssuer   = "https://www.servercontrolpanel.de/realms/scp"
	netcupClientId = "scp"
	netcupScopes   = []string{"offline_access", "openid"}
)

const refreshTokenTypeHint = "refresh_token"

type DefaultAuthenticator struct {
	authData *AuthData
	issuer   string
	clientId string
	scopes   []string
}

var _ Authenticator = &DefaultAuthenticator{}

// NewDefaultAuthenticator creates a new DefaultAuthenticator with the given access and refresh tokens.
// Refresh token can be empty, in which case new device authorization flow will be used.
func NewDefaultAuthenticator(accessToken, refreshToken string) *DefaultAuthenticator {
	return &DefaultAuthenticator{
		authData: &AuthData{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		},
		issuer:   netcupIssuer,
		clientId: netcupClientId,
		scopes:   netcupScopes,
	}
}

func (a *DefaultAuthenticator) createOIDCProvider(ctx context.Context) (rp.RelyingParty, error) {
	provider, err := rp.NewRelyingPartyOIDC(ctx, a.issuer, a.clientId, "", "", a.scopes)
	if err != nil {
		slog.Error("error creating OIDC provider", "error", err)
		return nil, err
	}
	return provider, nil
}

func (a *DefaultAuthenticator) newDeviceAuth(ctx context.Context, provider rp.RelyingParty) (*AuthData, error) {
	resp, err := rp.DeviceAuthorization(ctx, a.scopes, provider, nil)
	if err != nil {
		slog.Error("error during device authorization", "error", err)
		return nil, err
	}
	slog.Info("complete device authorization using given uri and code", "verification_uri", resp.VerificationURI, "user_code", resp.UserCode)
	token, err := rp.DeviceAccessToken(ctx, resp.DeviceCode, time.Duration(resp.Interval)*time.Second, provider)
	if err != nil {
		slog.Error("error getting access token", "error", err)
		return nil, err
	}
	if token.AccessToken == "" || token.RefreshToken == "" {
		slog.Error("received empty access token or refresh token during device authorization")
		return nil, err
	}
	slog.Debug("successfully obtained access token and refresh token via device authorization")
	return &AuthData{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
	}, nil
}

func (a *DefaultAuthenticator) refreshTokenAuth(ctx context.Context, provider rp.RelyingParty) (*AuthData, error) {
	token, err := rp.RefreshTokens[*oidc.IDTokenClaims](ctx, provider, a.authData.RefreshToken, "", "")
	if err != nil {
		slog.Error("error refreshing token", "error", err)
		return nil, err
	}
	slog.Debug("successfully refreshed access token using refresh token", "expiry", token.Expiry)
	return &AuthData{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		TokenType:    token.TokenType,
		Expiry:       token.Expiry,
		Subject:      token.IDTokenClaims.Subject,
	}, nil
}

func (a *DefaultAuthenticator) Authenticate(ctx context.Context) (*AuthResult, error) {
	provider, err := a.createOIDCProvider(ctx)
	if err != nil {
		return nil, err
	}

	// If refresh token is empty, use new device authorization flow.
	if a.authData.RefreshToken == "" {
		authData, err := a.newDeviceAuth(ctx, provider)

		// Update stored auth data.
		a.authData = authData

		return &AuthResult{
			IsNewDevice: true,
		}, err
	}
	// Otherwise, use refresh token flow for existing device.
	authData, err := a.refreshTokenAuth(ctx, provider)

	// Update stored auth data.
	a.authData = authData

	return &AuthResult{
		IsNewDevice: false,
	}, err
}

func (a *DefaultAuthenticator) Revoke(ctx context.Context) error {
	provider, err := a.createOIDCProvider(ctx)
	if err != nil {
		return err
	}

	if a.authData.RefreshToken == "" {
		return errors.New("no refresh token provided for revocation")
	}
	err = rp.RevokeToken(ctx, provider, a.authData.RefreshToken, refreshTokenTypeHint)
	if err != nil {
		slog.Error("error revoking refresh token", "error", err)
		return err
	}
	slog.Debug("successfully revoked refresh token")
	return nil
}

func (a *DefaultAuthenticator) GetUserInfo(ctx context.Context) (*UserInfo, error) {
	provider, err := rp.NewRelyingPartyOIDC(ctx, a.issuer, "", "", "", nil)
	if err != nil {
		slog.Error("error creating OIDC provider", "error", err)
		return nil, err
	}

	userInfo, err := rp.Userinfo[*oidc.UserInfo](ctx, a.authData.AccessToken, a.authData.TokenType, a.authData.Subject, provider)
	if err != nil {
		slog.Error("error getting user info", "error", err)
		return nil, err
	}
	slog.Debug("successfully obtained user info")
	return &UserInfo{userInfo}, nil
}

func (a *DefaultAuthenticator) IsAuthenticationExpired() bool {
	return time.Now().After(a.authData.Expiry)
}

func (a *DefaultAuthenticator) GetAuthData() *AuthData {
	return a.authData
}
