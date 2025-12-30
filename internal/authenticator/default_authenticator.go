package authenticator

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/zitadel/oidc/v3/pkg/client/rp"
	"github.com/zitadel/oidc/v3/pkg/oidc"
)

var (
	netcupIssuer   = "https://www.servercontrolpanel.de/realms/scp"
	netcupClientId = "scp"
	netcupScopes   = []string{"offline_access", "openid"}
)

type DefaultAuthenticator struct {
	AuthData *AuthData
	issuer   string
	clientId string
	scopes   []string
}

// NewDefaultAuthenticator creates a new DefaultAuthenticator with the given access and refresh tokens.
// Refresh token can be empty, in which case new device authorization flow will be used.
func NewDefaultAuthenticator(accessToken, refreshToken string) DefaultAuthenticator {
	return DefaultAuthenticator{
		AuthData: &AuthData{
			AccessToken:  accessToken,
			refreshToken: refreshToken,
		},
		issuer:   netcupIssuer,
		clientId: netcupClientId,
		scopes:   netcupScopes,
	}
}

func (a DefaultAuthenticator) newDeviceAuth(context context.Context, provider rp.RelyingParty) (*AuthData, error) {
	resp, err := rp.DeviceAuthorization(context, a.scopes, provider, nil)
	if err != nil {
		log.Fatalf("Error during device authorization: %v", err)
		return nil, err
	}
	fmt.Printf("\nPlease browse to %s and enter code %s\n", resp.VerificationURI, resp.UserCode)
	token, err := rp.DeviceAccessToken(context, resp.DeviceCode, time.Duration(resp.Interval)*time.Second, provider)
	if err != nil {
		log.Fatalf("Error getting access token: %v", err)
		return nil, err
	}
	fmt.Printf("Access Token: %s\n", token.AccessToken)
	fmt.Printf("Refresh Token: %s\n", token.RefreshToken)
	return &AuthData{
		AccessToken:  token.AccessToken,
		refreshToken: token.RefreshToken,
	}, nil
}

func (a DefaultAuthenticator) refreshTokenAuth(context context.Context, provider rp.RelyingParty) (*AuthData, error) {
	token, err := rp.RefreshTokens[*oidc.IDTokenClaims](context, provider, a.AuthData.refreshToken, "", "")
	if err != nil {
		log.Fatalf("Error refreshing token: %v", err)
		return nil, err
	}
	return &AuthData{
		AccessToken:  token.AccessToken,
		refreshToken: token.RefreshToken,
	}, nil
}

func (a DefaultAuthenticator) Authenticate(context context.Context) (*AuthData, error) {
	provider, err := rp.NewRelyingPartyOIDC(context, a.issuer, a.clientId, "", "", a.scopes)
	if err != nil {
		log.Fatalf("Error creating OIDC provider: %v", err)
	}
	if a.AuthData.refreshToken == "" {
		return a.newDeviceAuth(context, provider)
	}
	return a.refreshTokenAuth(context, provider)
}
