package types

import (
	"encoding/hex"
	"fmt"
	"reflect"

	"github.com/fxamacker/cbor/v2"
)

type Network byte

const (
	Testnet Network = 0
	Mainnet Network = 1
)

func (n Network) String() string {
	if n == Mainnet {
		return "mainnet"
	} else {
		return "testnet"
	}
}

type Coin uint64

type ProtocolParams struct {
	MinimumUtxoValue Coin
	PoolDeposit      uint64
	KeyDeposit       uint64
	MinFeeA          Coin
	MinFeeB          Coin
}

type AddrKeyHash = Hash28

type PoolKeyHash = Hash28

type Hash28 [28]byte

// NewHash28 creates a new Hash28 from a hex encoded string.
func NewHash28(h string) (Hash28, error) {
	hash := [28]byte{}
	b, err := hex.DecodeString(h)
	if err != nil {
		return hash, err
	}
	copy(hash[:], b)
	return hash, nil
}

// NewHash28FromBytes creates a new Hash28 from raw bytes.
func NewHash28FromBytes(b []byte) (Hash28, error) {
	hash := [28]byte{}
	if len(b) != 28 {
		return hash, fmt.Errorf("length should be 28")
	}
	copy(hash[:], b)
	return hash, nil
}

// String returns the hex encoding representation of a Hash28.
func (h Hash28) String() string {
	return hex.EncodeToString(h[:])
}

type Hash32 [32]byte

// NewHash32 creates a new Hash32 from a hex encoded string.
func NewHash32(h string) (Hash32, error) {
	hash := [32]byte{}
	b, err := hex.DecodeString(h)
	if err != nil {
		return hash, err
	}
	copy(hash[:], b)
	return hash, nil
}

// NewHash32FromBytes creates a new Hash32 from raw bytes.
func NewHash32FromBytes(b []byte) (Hash32, error) {
	hash := [32]byte{}
	if len(b) != 28 {
		return hash, fmt.Errorf("length should be 32")
	}
	copy(hash[:], b)
	return hash, nil
}

// String returns the hex encoding representation of a Hash32
func (h Hash32) String() string {
	return hex.EncodeToString(h[:])
}

type Uint64 *uint64

func NewUint64(u uint64) Uint64 {
	return Uint64(&u)
}

type String *string

func NewString(s string) String {
	return String(&s)
}

type UnitInterval = RationalNumber

type RationalNumber struct {
	_ struct{} `cbor:",toarray"`
	P uint64
	Q uint64
}

// MarshalCBOR implements cbor.Marshaler
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

// UnmarshalCBOR implements cbor.Unmarshaler
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
