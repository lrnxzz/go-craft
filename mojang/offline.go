package mojang

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
)

type Offline struct {
	Username string
}

func (o Offline) Authenticate(_ context.Context) (Session, error) {
	if o.Username == "" {
		return Session{}, fmt.Errorf("mojang: offline username is empty")
	}

	return Session{
		Profile: Profile{
			ID:   offlineUUID(o.Username),
			Name: o.Username,
		},
	}, nil
}

func offlineUUID(username string) string {
	hash := md5.Sum([]byte("OfflinePlayer:" + username))
	hash[6] = hash[6]&0x0f | 0x30
	hash[8] = hash[8]&0x3f | 0x80

	return hex.EncodeToString(hash[:])
}
