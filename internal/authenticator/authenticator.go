package authenticator

type AuthData struct {
	AccessToken  string
	refreshToken string
}

type Authenticator interface {
	Authenticate() (*AuthData, error)
}
