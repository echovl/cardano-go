package cardano

import (
	"github.com/echovl/bech32"
	"github.com/echovl/cardano-go/crypto"
	"golang.org/x/crypto/blake2b"
)

type Network byte

const (
	Testnet Network = 0
	Mainnet Network = 1
)

type Address string

// Bytes returns the byte slice representation of the address.
func (addr *Address) Bytes() []byte {
	_, bytes, err := bech32.DecodeToBase256(string(*addr))
	if err != nil {
		panic(err)
	}
	return bytes
}

func newEnterpriseAddress(xvk crypto.XVerificationKey, network Network) Address {
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

	hrp := getHrp(network)
	address, err := bech32.EncodeFromBase256(hrp, addressBytes)
	if err != nil {
		panic(err)
	}

	return Address(address)
}

// NewAddressFromBech32 creates an Address from a bech32 encoded string.
func NewAddressFromBech32(addr string) (Address, error) {
	_, _, err := bech32.DecodeToBase256(addr)
	if err != nil {
		return "", err
	}
	return Address(addr), nil
}

// NewAddressFromBytes creates an Address from a byte slice.
func NewAddressFromBytes(addr []byte, network Network) (Address, error) {
	encoded, err := bech32.EncodeFromBase256(getHrp(network), addr)
	if err != nil {
		return "", nil
	}
	return Address(encoded), nil
}

func getHrp(network Network) string {
	if network == Testnet {
		return "addr_test"
	} else {
		return "addr"
	}
}
