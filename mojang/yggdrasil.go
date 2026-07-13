package mojang

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/valyala/fasthttp"
)

const yggdrasilTimeout = 10 * time.Second

type yggdrasilAgent struct {
	Name    string `json:"name"`
	Version int    `json:"version"`
}

type yggdrasilRequest struct {
	Agent       yggdrasilAgent `json:"agent"`
	Username    string         `json:"username"`
	Password    string         `json:"password"`
	ClientToken string         `json:"clientToken,omitempty"`
	RequestUser bool           `json:"requestUser"`
}

type yggdrasilProfile struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type yggdrasilResponse struct {
	AccessToken     string           `json:"accessToken"`
	ClientToken     string           `json:"clientToken"`
	SelectedProfile yggdrasilProfile `json:"selectedProfile"`
}

type yggdrasilError struct {
	Err     string `json:"error"`
	Message string `json:"errorMessage"`
}

func (e *yggdrasilError) Error() string {
	return fmt.Sprintf("auth: yggdrasil %s: %s", e.Err, e.Message)
}

type Yggdrasil struct {
	BaseURL     string
	Email       string
	Password    string
	ClientToken string
}

func (y Yggdrasil) Authenticate(ctx context.Context) (Session, error) {
	body, err := json.Marshal(yggdrasilRequest{
		Agent: yggdrasilAgent{
			Name:    "Minecraft",
			Version: 1,
		},
		Username:    y.Email,
		Password:    y.Password,
		ClientToken: y.ClientToken,
		RequestUser: false,
	})
	if err != nil {
		return Session{}, err
	}

	request := fasthttp.AcquireRequest()
	response := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(request)
	defer fasthttp.ReleaseResponse(response)

	request.Header.SetMethod(fasthttp.MethodPost)
	request.SetRequestURI(y.BaseURL + "/authenticate")
	request.Header.SetContentType("application/json")
	request.Header.Set(fasthttp.HeaderAccept, "application/json")
	request.SetBody(body)

	deadline := time.Now().Add(yggdrasilTimeout)
	if ctxDeadline, ok := ctx.Deadline(); ok {
		deadline = ctxDeadline
	}

	if err := fasthttp.DoDeadline(request, response, deadline); err != nil {
		return Session{}, err
	}

	if response.StatusCode() != fasthttp.StatusOK {
		var failure yggdrasilError
		if err := json.Unmarshal(response.Body(), &failure); err == nil && failure.Err != "" {
			return Session{}, &failure
		}

		return Session{}, fmt.Errorf("auth: yggdrasil authenticate returned %d: %s", response.StatusCode(), response.Body())
	}

	var decoded yggdrasilResponse
	if err := json.Unmarshal(response.Body(), &decoded); err != nil {
		return Session{}, err
	}

	return Session{
		AccessToken: decoded.AccessToken,
		Profile: Profile{
			ID:   decoded.SelectedProfile.ID,
			Name: decoded.SelectedProfile.Name,
		},
	}, nil
}
