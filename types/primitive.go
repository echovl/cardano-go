package types

import (
	"reflect"

	"github.com/fxamacker/cbor/v2"
)

type Network byte

const (
	Testnet Network = 0
	Mainnet Network = 1
)

type ProtocolParams struct {
	MinimumUtxoValue Coin
	PoolDeposit      uint64
	KeyDeposit       uint64
	MinFeeA          Coin
	MinFeeB          Coin
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
