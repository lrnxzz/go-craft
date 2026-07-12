package auth_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/lrnxzz/go-craft/auth"
)

func TestYggdrasilAuthenticate(t *testing.T) {
	baseURL := os.Getenv("GOCRAFT_YGGDRASIL_URL")
	email := os.Getenv("GOCRAFT_YGGDRASIL_EMAIL")
	password := os.Getenv("GOCRAFT_YGGDRASIL_PASSWORD")

	if baseURL == "" || email == "" || password == "" {
		t.Skip("GOCRAFT_YGGDRASIL_URL, GOCRAFT_YGGDRASIL_EMAIL and GOCRAFT_YGGDRASIL_PASSWORD not set")
	}

	provider := auth.Yggdrasil{
		BaseURL:  baseURL,
		Email:    email,
		Password: password,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	session, err := provider.Authenticate(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if !session.Online() {
		t.Error("yggdrasil session carries no access token")
	}
	if session.Profile.Name == "" || len(session.Profile.ID) != 32 {
		t.Errorf("profile = %+v, want a name and 32-char id", session.Profile)
	}
}

func TestYggdrasilRejectsBadCredentials(t *testing.T) {
	baseURL := os.Getenv("GOCRAFT_YGGDRASIL_URL")
	if baseURL == "" {
		t.Skip("GOCRAFT_YGGDRASIL_URL not set")
	}

	provider := auth.Yggdrasil{
		BaseURL:  baseURL,
		Email:    "does-not-exist@example.com",
		Password: "wrong-password",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if _, err := provider.Authenticate(ctx); err == nil {
		t.Error("expected an error, got nil")
	}
}
