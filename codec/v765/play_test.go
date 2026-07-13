package v765_test

import (
	"reflect"
	"testing"

	gocraft "github.com/lrnxzz/go-craft"
	"github.com/lrnxzz/go-craft/codec/v765"
)

func TestJoinGameRoundTrip(t *testing.T) {
	original := &v765.JoinGame{
		EntityID:            42,
		Hardcore:            false,
		Worlds:              gocraft.Slice[gocraft.Identifier]{"minecraft:overworld", "minecraft:the_nether"},
		MaxPlayers:          20,
		ViewDistance:        10,
		SimulationDistance:  10,
		EnableRespawnScreen: true,
		DimensionType:       "minecraft:overworld",
		DimensionName:       "minecraft:overworld",
		HashedSeed:          -1234567890,
		GameMode:            0,
		PreviousGameMode:    -1,
		Flat:                true,
		Death: gocraft.Some(v765.DeathLocation{
			DimensionName: "minecraft:the_nether",
			Location:      gocraft.Position{X: 10, Y: 64, Z: -20},
		}),
		PortalCooldown: 0,
	}

	decoded := encodeAndDecode(t, gocraft.StatePlay, gocraft.Clientbound, original)

	if got := decoded.(*v765.JoinGame); !reflect.DeepEqual(got, original) {
		t.Errorf("got %+v, want %+v", got, original)
	}
}

func TestSyncPlayerPositionCarriesTeleportID(t *testing.T) {
	original := &v765.SyncPlayerPosition{
		X:          128.5,
		Y:          64.0,
		Z:          -256.25,
		Yaw:        90,
		Pitch:      -45,
		Flags:      0x1F,
		TeleportID: 7,
	}

	decoded := encodeAndDecode(t, gocraft.StatePlay, gocraft.Clientbound, original)

	got := decoded.(*v765.SyncPlayerPosition)
	if *got != *original {
		t.Errorf("got %+v, want %+v", got, original)
	}
	if got.TeleportID != 7 {
		t.Errorf("teleport id = %d, want 7 (the bot must echo this in ConfirmTeleport)", got.TeleportID)
	}
}

func TestPlayKeepAliveDistinctIDsPerDirection(t *testing.T) {
	var (
		clientbound v765.PlayKeepAlive
		serverbound v765.PlayKeepAliveResponse
	)

	if clientbound.ID() == serverbound.ID() {
		t.Fatal("clientbound and serverbound play keep-alive must have distinct ids")
	}

	received := &v765.PlayKeepAlive{KeepAliveID: 555}
	echoed := encodeAndDecode(t, gocraft.StatePlay, gocraft.Clientbound, received)
	if got := echoed.(*v765.PlayKeepAlive); got.KeepAliveID != 555 {
		t.Errorf("keep alive id = %d, want 555", got.KeepAliveID)
	}

	reply := &v765.PlayKeepAliveResponse{KeepAliveID: 555}
	confirmed := encodeAndDecode(t, gocraft.StatePlay, gocraft.Serverbound, reply)
	if got := confirmed.(*v765.PlayKeepAliveResponse); got.KeepAliveID != 555 {
		t.Errorf("keep alive reply id = %d, want 555", got.KeepAliveID)
	}
}

func TestConfirmTeleportRoundTrip(t *testing.T) {
	original := &v765.ConfirmTeleport{TeleportID: 7}

	decoded := encodeAndDecode(t, gocraft.StatePlay, gocraft.Serverbound, original)

	if got := decoded.(*v765.ConfirmTeleport); *got != *original {
		t.Errorf("got %+v, want %+v", got, original)
	}
}
