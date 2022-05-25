package cardano

import (
	"reflect"

	"github.com/echovl/bech32"
	"github.com/echovl/cardano-go/crypto"
	"github.com/fxamacker/cbor/v2"
	"golang.org/x/crypto/blake2b"
)

type Network byte

const (
	Testnet Network = 0
	Mainnet Network = 1
)

// Address is the bech32 representation of a cardano address
type Address string

type AddressBytes []byte

// Bytes returns the byte slice representation of the address.
func (addr *Address) Bytes() []byte {
	_, bytes, err := bech32.DecodeToBase256(string(*addr))
	if err != nil {
		panic(err)
	}
	return bytes
}

func DecodeAddress(data []byte) (Address, Address, error) {
	testnet, err := bech32.EncodeFromBase256("addr_test", data)
	if err != nil {
		return "", "", err
	}
	mainnet, err := bech32.EncodeFromBase256("addr", data)
	if err != nil {
		return "", "", err
	}
	return Address(mainnet), Address(testnet), nil
}

func NewEnterpriseAddress(xvk crypto.ExtendedVerificationKey, network Network) Address {
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

// Bech32ToAddress creates an Address from a bech32 encoded string.
func Bech32ToAddress(addr string) (Address, error) {
	_, _, err := bech32.DecodeToBase256(addr)
	if err != nil {
		return "", err
	}
	return Address(addr), nil
}

// BytesToAddress creates an Address from a byte slice.
func BytesToAddress(addr []byte, network Network) (Address, error) {
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

type Hash28 []byte

type Hash32 []byte

type AddrKeyHash Hash28

type PoolKeyHash Hash28

type Coin uint64

type Uint64 *uint64

func NewUint64(u uint64) Uint64 {
	return Uint64(&u)
}

type UnitInterval = RationalNumber

type RationalNumber struct {
	_ struct{} `cbor:",toarray"`
	P uint64
	Q uint64
}

func (r *RationalNumber) MarshalCBOR() ([]byte, error) {
	type rational RationalNumber

	// Register tag 30 for rational numbers
	tags, err := r.tagSet(rational{})
	if err != nil {
		return nil, err
	}

	em, err := cbor.EncOptions{}.EncModeWithTags(tags)
	if err != nil {
		return nil, err
	}

	return em.Marshal(rational(*r))
}

func (r *RationalNumber) UnmarshalCBOR(data []byte) error {
	type rational RationalNumber

	// Register tag 30 for rational numbers
	tags, err := r.tagSet(rational{})
	if err != nil {
		return err
	}

	dm, err := cbor.DecOptions{}.DecModeWithTags(tags)
	if err != nil {
		return err
	}

	var rr rational
	if err := dm.Unmarshal(data, &rr); err != nil {
		return err
	}
	r.P = rr.P
	r.Q = rr.Q

	return nil
}

func (r *RationalNumber) tagSet(contentType interface{}) (cbor.TagSet, error) {
	tags := cbor.NewTagSet()
	err := tags.Add(
		cbor.TagOptions{EncTag: cbor.EncTagRequired, DecTag: cbor.DecTagRequired},
		reflect.TypeOf(contentType),
		30)
	if err != nil {
		return nil, err
	}

	return tags, nil
}
