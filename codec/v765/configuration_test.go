package v765_test

import (
	"reflect"
	"testing"

	gocraft "github.com/lrnxzz/go-craft"
	"github.com/lrnxzz/go-craft/codec/v765"
	"github.com/lrnxzz/go-craft/nbt"
)

func TestClientInformationPreservesSettings(t *testing.T) {
	original := &v765.ClientInformation{
		Locale:              "en_us",
		ViewDistance:        12,
		ChatMode:            0,
		ChatColors:          true,
		DisplayedSkinParts:  0x7F,
		MainHand:            1,
		EnableTextFiltering: false,
		EnableServerListing: true,
	}

	proto := v765.Protocol()
	decoded, ok, err := proto.Decode(gocraft.StateConfiguration, gocraft.Serverbound, gocraft.EncodeFrame(original))
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("client information not registered")
	}

	if got := decoded.(*v765.ClientInformation); *got != *original {
		t.Errorf("got %+v, want %+v", got, original)
	}
}

func TestConfigKeepAliveDecodesInBothDirections(t *testing.T) {
	original := &v765.ConfigKeepAlive{
		KeepAliveID: 1234567890,
	}

	for _, dir := range []gocraft.Direction{gocraft.Clientbound, gocraft.Serverbound} {
		proto := v765.Protocol()
		decoded, ok, err := proto.Decode(gocraft.StateConfiguration, dir, gocraft.EncodeFrame(original))
		if err != nil {
			t.Fatalf("%s: %v", dir, err)
		}
		if !ok {
			t.Fatalf("%s: keep-alive not registered", dir)
		}

		if got := decoded.(*v765.ConfigKeepAlive); *got != *original {
			t.Errorf("%s: got %+v, want %+v", dir, got, original)
		}
	}
}

func TestConfigDisconnectCarriesNBTReason(t *testing.T) {
	original := &v765.ConfigDisconnect{
		Reason: gocraft.NBT{
			"text": nbt.String("kicked"),
		},
	}

	proto := v765.Protocol()
	decoded, ok, err := proto.Decode(gocraft.StateConfiguration, gocraft.Clientbound, gocraft.EncodeFrame(original))
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("config disconnect not registered")
	}

	if got := decoded.(*v765.ConfigDisconnect); !reflect.DeepEqual(got.Reason, original.Reason) {
		t.Errorf("reason = %#v, want %#v", got.Reason, original.Reason)
	}
}

func TestRegistryDataCarriesNBTCodec(t *testing.T) {
	original := &v765.RegistryData{
		Codec: gocraft.NBT{
			"minecraft:dimension_type": nbt.Compound{
				"type": nbt.String("minecraft:dimension_type"),
			},
		},
	}

	proto := v765.Protocol()
	decoded, ok, err := proto.Decode(gocraft.StateConfiguration, gocraft.Clientbound, gocraft.EncodeFrame(original))
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("registry data not registered")
	}

	if got := decoded.(*v765.RegistryData); !reflect.DeepEqual(got.Codec, original.Codec) {
		t.Errorf("codec = %#v, want %#v", got.Codec, original.Codec)
	}
}

func TestFeatureFlagsCarriesFeatureList(t *testing.T) {
	original := &v765.FeatureFlags{
		Features: gocraft.Slice[gocraft.Identifier]{"minecraft:vanilla", "minecraft:bundle"},
	}

	proto := v765.Protocol()
	decoded, ok, err := proto.Decode(gocraft.StateConfiguration, gocraft.Clientbound, gocraft.EncodeFrame(original))
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("feature flags not registered")
	}

	if got := decoded.(*v765.FeatureFlags); !reflect.DeepEqual(got, original) {
		t.Errorf("got %+v, want %+v", got, original)
	}
}
