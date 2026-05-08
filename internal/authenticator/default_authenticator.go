package authenticator

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"golang.org/x/oauth2"
)

var (
	netcupAuthUrl       = "https://www.servercontrolpanel.de/realms/scp/protocol/openid-connect/auth"
	netcupTokenUrl      = "https://www.servercontrolpanel.de/realms/scp/protocol/openid-connect/token"
	netcupDeviceAuthUrl = "https://www.servercontrolpanel.de/realms/scp/protocol/openid-connect/auth/device"
	netcupClientId      = "scp"
	netcupScopes        = []string{"offline_access", "openid"}
)

type DefaultAuthenticator struct {
	refreshToken        *string
	authenticatedClient *http.Client
	clientId            string
	scopes              []string
}

var _ Authenticator = &DefaultAuthenticator{}

// NewDefaultAuthenticator creates a new DefaultAuthenticator with the given refresh token.
// Refresh token can be empty, in which case new device authorization flow will be used.
func NewDefaultAuthenticator(refreshToken string) *DefaultAuthenticator {
	return &DefaultAuthenticator{
		refreshToken: &refreshToken,
		clientId:     netcupClientId,
		scopes:       netcupScopes,
	}
}

func (a *DefaultAuthenticator) createOAuthConfig() (*oauth2.Config, error) {
	config := &oauth2.Config{
		ClientID: a.clientId,
		Scopes:   a.scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:       netcupAuthUrl,
			TokenURL:      netcupTokenUrl,
			DeviceAuthURL: netcupDeviceAuthUrl,
		},
	}
	return config, nil
}

func (a *DefaultAuthenticator) newDeviceAuth(ctx context.Context, oauthConfig *oauth2.Config) (string, error) {
	deviceAuth, err := oauthConfig.DeviceAuth(ctx)
	if err != nil {
		slog.Error("error during device authorization", "error", err)
		return "", err
	}
	slog.Info("complete device authorization using given uri and code", "verification_uri", deviceAuth.VerificationURI, "user_code", deviceAuth.UserCode)
	token, err := oauthConfig.DeviceAccessToken(ctx, deviceAuth)
	if err != nil {
		slog.Error("error getting access token", "error", err)
		return "", err
	}
	if token.AccessToken == "" || token.RefreshToken == "" {
		slog.Error("received empty access token or refresh token during device authorization")
		return "", errors.New("received empty access token or refresh token during device authorization")
	}
	slog.Debug("successfully obtained access token and refresh token via device authorization")
	return token.RefreshToken, nil
}

func (a *DefaultAuthenticator) refreshTokenAuth(ctx context.Context, oauthConfig *oauth2.Config) error {
	token := &oauth2.Token{
		RefreshToken: *a.refreshToken,
	}
	client := oauthConfig.Client(ctx, token)
	a.authenticatedClient = client
	slog.Debug("successfully obtained authenticated client using refresh token")
	return nil
}

func (a *DefaultAuthenticator) Authenticate(ctx context.Context) (*AuthResult, error) {
	oauthConfig, err := a.createOAuthConfig()
	if err != nil {
		return nil, err
	}

	// If refresh token is empty, use new device authorization flow.
	if a.refreshToken == nil || *a.refreshToken == "" {
		refreshToken, err := a.newDeviceAuth(ctx, oauthConfig)
		if err != nil {
			return nil, err
		}

		return &AuthResult{
			IsNewDevice:  true,
			RefreshToken: refreshToken,
		}, nil
	}
	// Otherwise, use refresh token flow for existing device.
	err = a.refreshTokenAuth(ctx, oauthConfig)
	if err != nil {
		return nil, err
	}

	return &AuthResult{
		IsNewDevice: false,
	}, nil
}

func (a *DefaultAuthenticator) GetAuthenticatedClient() *http.Client {
	return a.authenticatedClient
}
