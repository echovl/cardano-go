package wallet

import (
	"github.com/echovl/bech32"
	"github.com/echovl/cardano-wallet/crypto"
	"golang.org/x/crypto/blake2b"
)

type Network byte

const (
	Testnet Network = 0
	Mainnet Network = 1
)

type Address string

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

func getHrp(network Network) string {
	if network == Testnet {
		return "addr_test"
	} else {
		return "addr"
	}
}
