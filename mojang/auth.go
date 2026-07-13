package mojang

import "context"

type Session struct {
	AccessToken string
	Profile     Profile
}

func (s Session) Online() bool {
	return s.AccessToken != ""
}

type Authenticator interface {
	Authenticate(ctx context.Context) (Session, error)
}
