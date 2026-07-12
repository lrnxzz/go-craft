package auth_test

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"os"
	"testing"
	"time"

	"github.com/lrnxzz/go-craft/auth"
)

func _sessionToken(t *testing.T) string {
	t.Helper()

	token := os.Getenv("GOCRAFT_ACCESS_TOKEN")
	if token == "" {
		t.Skip("GOCRAFT_ACCESS_TOKEN not set")
	}

	return token
}

func TestMojangProfile(t *testing.T) {
	mojang := auth.NewMojang(_sessionToken(t))

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	profile, err := mojang.Profile(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if len(profile.ID) != 32 {
		t.Errorf("profile id = %q, want 32 hex chars", profile.ID)
	}
	if profile.Name == "" {
		t.Error("profile name is empty")
	}
}

func TestMojangProfileWithInvalidToken(t *testing.T) {
	_sessionToken(t)

	mojang := auth.NewMojang("invalid-token")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if _, err := mojang.Profile(ctx); err == nil {
		t.Error("expected an error, got nil")
	}
}

func TestMojangJoinAndHasJoined(t *testing.T) {
	mojang := auth.NewMojang(_sessionToken(t))

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	profile, err := mojang.Profile(ctx)
	if err != nil {
		t.Fatal(err)
	}

	nonce := make([]byte, 20)
	if _, err := rand.Read(nonce); err != nil {
		t.Fatal(err)
	}

	serverID := hex.EncodeToString(nonce)

	if err := mojang.JoinServer(ctx, profile, serverID); err != nil {
		t.Fatal(err)
	}

	joined, err := mojang.HasJoined(ctx, profile.Name, serverID)
	if err != nil {
		t.Fatal(err)
	}

	if joined.ID != profile.ID {
		t.Errorf("hasJoined returned id %q, want %q", joined.ID, profile.ID)
	}
}
