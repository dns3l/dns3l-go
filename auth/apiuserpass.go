package auth

import (
	"context"
	"net/url"

	"golang.org/x/oauth2"
)

// AuthAPIUserPass logs you in with a username/password-based Resource Owner Password
// Credentials (ROPC) grant. Unless your dex is configured differently to the official
// dns3l stack, the username and password are your LDAP credentials.
func AuthAPIUserPass(ctx context.Context, instance, username, password string) (
	*oauth2.Token, error) {

	urlauth, err := url.JoinPath(instance, "/auth/token")
	if err != nil {
		return nil, err
	}

	cfg := &oauth2.Config{
		Endpoint: oauth2.Endpoint{
			TokenURL: urlauth,
		},
		ClientID: "dns3l-api-pub",
		Scopes: []string{
			"openid",
			"profile",
			"groups",
			"email",
			"audience:server:client_id:dns3ld",
		},
	}

	token, err := cfg.PasswordCredentialsToken(
		ctx,
		username,
		password,
	)
	if err != nil {
		return nil, err
	}

	return token, err
}
