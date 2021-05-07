package wallet

import (
	"github.com/echovl/bech32"
	"github.com/echovl/cardano-wallet/crypto"
	"golang.org/x/crypto/blake2b"
)

const (
	Testnet NetworkType = 0
	Mainnet NetworkType = 1
)

type NetworkType byte
type Address string

func newEnterpriseAddress(xvk crypto.XVerificationKey, network NetworkType) (Address, error) {
	addressBytes := make([]byte, 29)
	header := 0x60 | (byte(network) & 0xFF)
	hash, err := blake2b.New(224/8, nil)
	if err != nil {
		return "", err
	}

	hash.Write(xvk[:32])
	paymentHash := hash.Sum(nil)

	addressBytes[0] = header
	copy(addressBytes[1:], paymentHash)

	address, err := bech32.EncodeFromBase256("addr_test", addressBytes)
	if err != nil {
		return "", err
	}

	return Address(address), nil
}
