package bech32

import (
	"fmt"

	"github.com/echovl/cardano-go/bech32/prefixes"
	"github.com/echovl/cardano-go/internal/bech32"
)

type Bech32Prefix = prefixes.Bech32Prefix

var (
	EncodeWithPrefix = bech32.Encode
	Decode           = bech32.Decode
	DecodeNoLimit    = bech32.DecodeNoLimit

	EncodeFromBase256WithPrefix = bech32.EncodeFromBase256
	DecodeToBase256             = bech32.DecodeToBase256
)

type (
	Bech32Codec interface {
		Prefix() string
		Bytes() []byte
		SetBytes([]byte)
		Len() int
	}
	Bech32Encoder interface {
		Prefix() string
		Bytes() []byte
	}
)

func Encode(args ...any) (string, error) {
	var hrp string
	var data []byte
	switch len(args) {
	case 1:
		// the argument have to be a Bech32Codec
		if c, ok := args[0].(Bech32Encoder); !ok {
			return "", fmt.Errorf("Wrong parameter: %T is not a Bech32Encoder", c)
		}
		hrp = args[0].(Bech32Encoder).Prefix()
		data = args[0].(Bech32Encoder).Bytes()
	case 2:
		// the argument haave to be a Bech32Codec or a string, and the second have to be a []byte
		a1 := args[0]
		a2 := args[1]

		switch a1.(type) {
		case Bech32Encoder:
			hrp = a1.(Bech32Codec).Prefix()
		case string:
			hrp = a1.(string)
		default:
			return "", fmt.Errorf("Wrong 1st parameter: %T is not a string or Bech32Codec", a1)
		}

		if _, ok := a2.([]byte); !ok {
			return "", fmt.Errorf("Wrong 2nd parameter: %T is not a []byte", a2)
		}
		data = a2.([]byte)
	}
	return bech32.Encode(hrp, data)
}

func EncodeFromBase256(args ...any) (string, error) {
	var hrp string
	var data []byte
	switch len(args) {
	case 1:
		// the argument have to be a Bech32Codec
		if c, ok := args[0].(Bech32Encoder); !ok {
			return "", fmt.Errorf("Wrong parameter: %T is not a Bech32Encoder", c)
		}
		hrp = args[0].(Bech32Encoder).Prefix()
		if converted, err := bech32.ConvertBits(args[0].(Bech32Encoder).Bytes(), 8, 5, true); err != nil {
			return "", err
		} else {
			data = converted
		}
	case 2:
		// the argument haave to be a Bech32Codec or a string, and the second have to be a []byte
		a1 := args[0]
		a2 := args[1]

		switch a1.(type) {
		case Bech32Encoder:
			hrp = a1.(Bech32Encoder).Prefix()
		case string:
			hrp = a1.(string)
		default:
			return "", fmt.Errorf("Wrong 1st parameter: %T is not a string or Bech32Codec", a1)
		}

		if _, ok := a2.([]byte); !ok {
			return "", fmt.Errorf("Wrong 2nd parameter: %T is not a []byte", a2)
		}
		if converted, err := bech32.ConvertBits(a2.([]byte), 8, 5, true); err != nil {
		} else {
			data = converted
		}
	}
	return bech32.Encode(hrp, data)
}

func DecodeInto(be32 string, codec Bech32Codec) error {
	expectedLen := codec.Len()
	expectedPrefix := codec.Prefix()
	hrp, data, err := bech32.DecodeToBase256(be32)
	if err != nil {
		return err
	}
	if hrp != expectedPrefix {
		return fmt.Errorf("Wrong prefix: want %s got %s", expectedPrefix, hrp)
	}
	codec.SetBytes(data)
	if len(codec.Bytes()) != expectedLen || string(codec.Bytes()) != string(data) {
		return fmt.Errorf("Set bytes failed")
	}
	return nil
}
