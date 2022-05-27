package types

import (
	"github.com/echovl/bech32"
	"github.com/echovl/cardano-go/crypto"
	"github.com/fxamacker/cbor/v2"
	"golang.org/x/crypto/blake2b"
)

// Address is a cardano address
type Address struct {
	B   []byte
	Hrp string
}

// NewAddress creates an Address from a bech32 encoded string.
func NewAddress(addr string) (Address, error) {
	hrp, bytes, err := bech32.DecodeToBase256(addr)
	if err != nil {
		return Address{}, err
	}
	return Address{B: bytes, Hrp: hrp}, nil
}

func (a Address) String() string {
	addr, err := bech32.EncodeFromBase256(a.Hrp, a.B)
	if err != nil {
		panic(err)
	}

	return addr
}

// MarshalCBOR implements cbor.Marshaler
func (a *Address) MarshalCBOR() ([]byte, error) {
	return cbor.Marshal(a.B)
}

// UnmarshalCBOR implements cbor.Unmarshaler
func (a *Address) UnmarshalCBOR(data []byte) error {
	if err := cbor.Unmarshal(data, &a.B); err != nil {
		return nil
	}
	a.Hrp = getHrp(Network(a.B[0] & 0x01))
	return nil
}

func NewEnterpriseAddress(xvk crypto.XPub, network Network) Address {
	addressBytes := make([]byte, 29)
	header := 0x60 | (byte(network) & 0xFF)
	hash, err := blake2b.New(224/8, nil)
	if err != nil {
		panic(err)
	}

	hash.Write(xvk[:32])
	paymentHash := hash.Sum(nil)

	addressBytes[0] = header
	copy(addressBytes[1:], paymentHash)

	return Address{B: addressBytes, Hrp: getHrp(network)}
}

func getHrp(network Network) string {
	if network == Testnet {
		return "addr_test"
	} else {
		return "addr"
	}
}
