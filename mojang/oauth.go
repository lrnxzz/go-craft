package mojang

import (
	"fmt"
	"net/url"
	"time"
)

const (
	fallbackPollInterval = 5 * time.Second
	slowDownPenalty      = 5 * time.Second
)

type Flow struct {
	DeviceCodeURL string
	TokenURL      string
	Scope         string
	ResponseType  string
	XboxPreamble  string
}

var (
	Live = Flow{
		DeviceCodeURL: "https://login.live.com/oauth20_connect.srf",
		TokenURL:      "https://login.live.com/oauth20_token.srf",
		Scope:         "service::user.auth.xboxlive.com::MBI_SSL",
		ResponseType:  "device_code",
		XboxPreamble:  "t=",
	}

	Entra = Flow{
		DeviceCodeURL: "https://login.microsoftonline.com/consumers/oauth2/v2.0/devicecode",
		TokenURL:      "https://login.microsoftonline.com/consumers/oauth2/v2.0/token",
		Scope:         "XboxLive.signin offline_access",
		XboxPreamble:  "d=",
	}
)

type grantType string

const (
	grantDeviceCode grantType = "urn:ietf:params:oauth:grant-type:device_code"
	grantRefresh    grantType = "refresh_token"
)

type oauthErrorCode string

const (
	oauthAuthorizationPending oauthErrorCode = "authorization_pending"
	oauthSlowDown             oauthErrorCode = "slow_down"
)

type OAuthError struct {
	Code        oauthErrorCode `json:"error"`
	Description string         `json:"error_description"`
}

func (e *OAuthError) Error() string {
	return fmt.Sprintf("auth: oauth failure %s: %s", e.Code, e.Description)
}

type DeviceCode struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationURI string `json:"verification_uri"`
	ExpiresIn       int    `json:"expires_in"`
	Interval        int    `json:"interval"`
	Message         string `json:"message"`
}

func (c DeviceCode) PollInterval() time.Duration {
	if c.Interval <= 0 {
		return fallbackPollInterval
	}

	return time.Duration(c.Interval) * time.Second
}

func (c DeviceCode) Lifetime() time.Duration {
	return time.Duration(c.ExpiresIn) * time.Second
}

type TokenSet struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

func (t TokenSet) Lifetime() time.Duration {
	return time.Duration(t.ExpiresIn) * time.Second
}

type msaForm struct {
	clientID     string
	grant        grantType
	deviceCode   string
	refreshToken string
	scope        string
	responseType string
}

func (f msaForm) encode() string {
	form := url.Values{}
	form.Set("client_id", f.clientID)

	if f.grant != "" {
		form.Set("grant_type", string(f.grant))
	}
	if f.deviceCode != "" {
		form.Set("device_code", f.deviceCode)
	}
	if f.refreshToken != "" {
		form.Set("refresh_token", f.refreshToken)
	}
	if f.scope != "" {
		form.Set("scope", f.scope)
	}
	if f.responseType != "" {
		form.Set("response_type", f.responseType)
	}

	return form.Encode()
}
