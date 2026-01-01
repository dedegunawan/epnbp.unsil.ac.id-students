package authoidc

import (
	"context"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/dedegunawan/epnbp.unsil.ac.id-students-backend2/pkg/logger"
	"golang.org/x/oauth2"
	"net/url"
)

var Provider *oidc.Provider
var Verifier *oidc.IDTokenVerifier
var OAuth2Config *oauth2.Config
var LogoutEndpoint string

type AuthOidc struct {
	Issuer         string
	ClientID       string
	ClientSecret   string
	RedirectURI    string
	LogoutRedirect string
	LogoutEndpoint string
	logger         *logger.Logger
	Provider       *oidc.Provider
	Verifier       *oidc.IDTokenVerifier
	OAuth2Config   *oauth2.Config
}

func NewAuthOidc(issuer string, clientID string, clientSecret string, redirectURI string, logoutRedirect string, logoutEndpoint string, lg *logger.Logger) (*AuthOidc, error) {
	authOidc := &AuthOidc{
		Issuer:         issuer,
		ClientID:       clientID,
		ClientSecret:   clientSecret,
		RedirectURI:    redirectURI,
		LogoutRedirect: logoutRedirect,
		LogoutEndpoint: logoutEndpoint,
		logger:         lg,
	}
	err := authOidc.InitOIDC()
	return authOidc, err
}

func (a *AuthOidc) InitOIDC() error {
	ctx := context.Background()

	provider, err := oidc.NewProvider(ctx, a.Issuer)
	if err != nil {
		a.logger.Fatal("❌ Failed to init Keycloak OIDC provider:", err)
	}
	a.Provider = provider

	a.Verifier = provider.Verifier(&oidc.Config{
		ClientID:          a.ClientID,
		SkipClientIDCheck: true,
	})

	a.OAuth2Config = &oauth2.Config{
		ClientID:     a.ClientID,
		ClientSecret: a.ClientSecret,
		RedirectURL:  a.RedirectURI,
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}

	// Ambil metadata tambahan (termasuk logout endpoint)
	var metadata struct {
		EndSessionEndpoint string `json:"end_session_endpoint"`
	}
	if err := provider.Claims(&metadata); err != nil {
		return err
	}
	a.LogoutEndpoint = metadata.EndSessionEndpoint
	return nil
}

func (a *AuthOidc) GetLogoutURL() string {

	if a.LogoutEndpoint == "" {
		return "" // fallback manual kalau perlu
	}
	params := url.Values{}
	params.Add("post_logout_redirect_uri", a.LogoutRedirect)
	params.Add("client_id", a.ClientID)

	return a.LogoutEndpoint + "?" + params.Encode()
}
