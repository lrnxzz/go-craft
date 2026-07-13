package mojang

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/valyala/fasthttp"
)

const (
	loginWithXboxURL = "https://api.minecraftservices.com/authentication/login_with_xbox"

	loginTimeout = 10 * time.Second
)

type identityRequest struct {
	IdentityToken string `json:"identityToken"`
}

type minecraftToken struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

type Microsoft struct {
	ClientID string
	Prompt   func(DeviceCode)
	Flow     Flow
}

func (m Microsoft) Authenticate(ctx context.Context) (Session, error) {
	flow := m.Flow
	if flow.DeviceCodeURL == "" {
		flow = Live()
	}

	msa := NewMSA(m.ClientID)
	msa.Flow = flow

	code, err := msa.RequestDeviceCode(ctx)
	if err != nil {
		return Session{}, err
	}

	if m.Prompt != nil {
		m.Prompt(code)
	}

	tokens, err := msa.AwaitToken(ctx, code)
	if err != nil {
		return Session{}, err
	}

	xbox := NewXbox()
	xbox.Preamble = flow.XboxPreamble

	identity, err := xbox.Identity(ctx, tokens.AccessToken)
	if err != nil {
		return Session{}, err
	}

	minecraft, err := loginWithIdentity(ctx, identity)
	if err != nil {
		return Session{}, err
	}

	profile, err := NewMojang(minecraft.AccessToken).Profile(ctx)
	if err != nil {
		return Session{}, err
	}

	return Session{
		AccessToken: minecraft.AccessToken,
		Profile:     profile,
	}, nil
}

func loginWithIdentity(ctx context.Context, identityToken string) (minecraftToken, error) {
	body, err := json.Marshal(identityRequest{
		IdentityToken: identityToken,
	})
	if err != nil {
		return minecraftToken{}, err
	}

	request := fasthttp.AcquireRequest()
	response := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(request)
	defer fasthttp.ReleaseResponse(response)

	request.Header.SetMethod(fasthttp.MethodPost)
	request.SetRequestURI(loginWithXboxURL)
	request.Header.SetContentType("application/json")
	request.Header.Set(fasthttp.HeaderAccept, "application/json")
	request.SetBody(body)

	deadline := time.Now().Add(loginTimeout)
	if ctxDeadline, ok := ctx.Deadline(); ok {
		deadline = ctxDeadline
	}

	if err := fasthttp.DoDeadline(request, response, deadline); err != nil {
		return minecraftToken{}, err
	}
	if response.StatusCode() != fasthttp.StatusOK {
		return minecraftToken{}, fmt.Errorf("auth: login_with_xbox returned %d: %s", response.StatusCode(), response.Body())
	}

	var token minecraftToken
	if err := json.Unmarshal(response.Body(), &token); err != nil {
		return minecraftToken{}, err
	}

	return token, nil
}
