package auth

import (
	"context"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/dedegunawan/backend-ujian-telp-v5/utils"
	"golang.org/x/oauth2"
	"net/url"
	"os"
)

var Provider *oidc.Provider
var Verifier *oidc.IDTokenVerifier
var OAuth2Config *oauth2.Config
var LogoutEndpoint string

func InitOIDC() {
	ctx := context.Background()
	issuer := os.Getenv("OIDC_ISSUER") // e.g. http://localhost:8080/realms/myrealm

	provider, err := oidc.NewProvider(ctx, issuer)
	if err != nil {
		utils.Log.Fatal("❌ Failed to init Keycloak OIDC provider:", err)
	}
	Provider = provider

	Verifier = provider.Verifier(&oidc.Config{
		ClientID:          os.Getenv("OIDC_CLIENT_ID"),
		SkipClientIDCheck: true,
	})

	OAuth2Config = &oauth2.Config{
		ClientID:     os.Getenv("OIDC_CLIENT_ID"),
		ClientSecret: os.Getenv("OIDC_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("OIDC_REDIRECT_URI"),
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}

	// Ambil metadata tambahan (termasuk logout endpoint)
	var metadata struct {
		EndSessionEndpoint string `json:"end_session_endpoint"`
	}
	if err := provider.Claims(&metadata); err != nil {
		utils.Log.Warnf("⚠️ Failed to read provider claims (logout endpoint may not work): %v", err)
		LogoutEndpoint = "" // Set empty, GetLogoutURL will handle it
	} else {
		LogoutEndpoint = metadata.EndSessionEndpoint
	}
}

func GetLogoutURL(postLogoutRedirectURI string, clientID string) string {
	if LogoutEndpoint == "" {
		return "" // fallback manual kalau perlu
	}
	params := url.Values{}
	params.Add("post_logout_redirect_uri", postLogoutRedirectURI)
	params.Add("client_id", clientID)

	return LogoutEndpoint + "?" + params.Encode()
}
