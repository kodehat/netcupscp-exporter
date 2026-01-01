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
	AuthData *AuthData
	issuer   string
	clientId string
	scopes   []string
}

var _ Authenticator = DefaultAuthenticator{}

// NewDefaultAuthenticator creates a new DefaultAuthenticator with the given access and refresh tokens.
// Refresh token can be empty, in which case new device authorization flow will be used.
func NewDefaultAuthenticator(accessToken, refreshToken string) DefaultAuthenticator {
	return DefaultAuthenticator{
		AuthData: &AuthData{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		},
		issuer:   netcupIssuer,
		clientId: netcupClientId,
		scopes:   netcupScopes,
	}
}

func (a DefaultAuthenticator) newDeviceAuth(context context.Context, provider rp.RelyingParty) (*AuthData, error) {
	resp, err := rp.DeviceAuthorization(context, a.scopes, provider, nil)
	if err != nil {
		slog.Error("error during device authorization", "error", err)
		return nil, err
	}
	slog.Info("complete device authorization using given uri and code", "verification_uri", resp.VerificationURI, "user_code", resp.UserCode)
	token, err := rp.DeviceAccessToken(context, resp.DeviceCode, time.Duration(resp.Interval)*time.Second, provider)
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

func (a DefaultAuthenticator) refreshTokenAuth(context context.Context, provider rp.RelyingParty) (*AuthData, error) {
	token, err := rp.RefreshTokens[*oidc.IDTokenClaims](context, provider, a.AuthData.RefreshToken, "", "")
	if err != nil {
		slog.Error("error refreshing token", "error", err)
		return nil, err
	}
	slog.Debug("successfully refreshed access token using refresh token")
	return &AuthData{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
	}, nil
}

func (a DefaultAuthenticator) Authenticate(context context.Context) (*AuthResult, error) {
	provider, err := rp.NewRelyingPartyOIDC(context, a.issuer, a.clientId, "", "", a.scopes)
	if err != nil {
		slog.Error("error creating OIDC provider", "error", err)
	}
	// If refresh token is empty, use new device authorization flow.
	if a.AuthData.RefreshToken == "" {
		authData, err := a.newDeviceAuth(context, provider)
		return &AuthResult{
			AuthData:    authData,
			IsNewDevice: true,
		}, err
	}
	// Otherwise, use refresh token flow for existing device.
	authData, err := a.refreshTokenAuth(context, provider)
	return &AuthResult{
		AuthData:    authData,
		IsNewDevice: false,
	}, err
}

func (a DefaultAuthenticator) Revoke(ctx context.Context) error {
	provider, err := rp.NewRelyingPartyOIDC(ctx, a.issuer, a.clientId, "", "", a.scopes)
	if err != nil {
		slog.Error("error creating OIDC provider", "error", err)
		return err
	}
	if a.AuthData.RefreshToken == "" {
		return errors.New("no refresh token provided for revocation")
	}
	err = rp.RevokeToken(ctx, provider, a.AuthData.RefreshToken, refreshTokenTypeHint)
	if err != nil {
		slog.Error("error revoking refresh token", "error", err)
		return err
	}
	slog.Debug("successfully revoked refresh token")
	return nil
}
