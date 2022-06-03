package cardano

import (
	"encoding/hex"
	"errors"
	"reflect"

	"github.com/echovl/cardano-go/internal/cbor"
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

type BigNum uint64

type Coin BigNum

type Value struct {
	Coin       Coin
	MultiAsset *MultiAsset
}

func NewValue(coin Coin) *Value {
	return &Value{Coin: coin, MultiAsset: NewMultiAsset()}
}

func NewValueWithAssets(coin Coin, assets *MultiAsset) *Value {
	return &Value{Coin: coin, MultiAsset: assets}
}

func (v *Value) OnlyCoin() bool {
	return v == nil || len(v.MultiAsset.m) == 0
}

func (v *Value) IsZero() bool {
	for _, assets := range v.MultiAsset.m {
		for _, value := range assets.m {
			if value != 0 {
				return false
			}
		}
	}
	return v.Coin == 0
}

func (v *Value) Add(rhs *Value) *Value {
	coin := v.Coin + rhs.Coin
	result := NewValue(coin)

	for _, ma := range []*MultiAsset{v.MultiAsset, rhs.MultiAsset} {
		for policy, assets := range ma.m {
			for assetName, value := range assets.m {
				current := result.MultiAsset
				if _, policyExists := current.m[policy]; policyExists {
					if _, assetExists := result.MultiAsset.m[policy].m[assetName]; assetExists {
						current.m[policy].m[assetName] += value
					} else {
						current.m[policy].m[assetName] = value
					}
				} else {
					current.m[policy] = &Assets{
						m: map[cbor.ByteString]BigNum{
							assetName: value,
						},
					}
				}
			}
		}
	}

	return result
}

func (v *Value) Sub(rhs *Value) *Value {
	var coin Coin
	if v.Coin > rhs.Coin {
		coin = v.Coin - rhs.Coin
	}

	result := NewValue(coin)
	for policy, assets := range v.MultiAsset.m {
		result.MultiAsset.m[policy] = assets
	}

	for policy, assets := range rhs.MultiAsset.m {
		for assetName, value := range assets.m {
			current := result.MultiAsset
			if _, policyExists := current.m[policy]; policyExists {
				if _, assetExists := result.MultiAsset.m[policy].m[assetName]; assetExists {
					lastValue := current.m[policy].m[assetName]
					if lastValue > lastValue-value {
						current.m[policy].m[assetName] -= value
					} else {
						current.m[policy].m[assetName] = 0
					}
				} else {
					current.m[policy].m[assetName] = 0
				}
			} else {
				current.m[policy] = &Assets{
					m: map[cbor.ByteString]BigNum{
						assetName: 0,
					},
				}
			}
		}
	}

	return result
}

func (v *Value) Cmp(rhs *Value) (int, error) {
	lrZero := v.Sub(rhs).IsZero()
	rlZero := rhs.Sub(v).IsZero()

	if !lrZero && !rlZero {
		return 0, errors.New("noncomparable values")
	} else if !lrZero && rlZero {
		return -1, nil
	} else if lrZero && !rlZero {
		return 1, nil
	} else {
		return 0, nil
	}
}

// MarshalCBOR implements cbor.Marshaler
func (v *Value) MarshalCBOR() ([]byte, error) {
	if v.OnlyCoin() {
		return cborEnc.Marshal(v.Coin)
	} else {
		return cborEnc.Marshal([]interface{}{v.Coin, v.MultiAsset})
	}
}

type PolicyID struct {
	bs cbor.ByteString
}

func NewPolicyID(script NativeScript) (PolicyID, error) {
	scriptHash, err := script.Hash()
	if err != nil {
		return PolicyID{}, err
	}
	return PolicyID{bs: cbor.NewByteString(scriptHash)}, nil
}

func NewPolicyIDFromHash(scriptHash Hash28) PolicyID {
	return PolicyID{bs: cbor.NewByteString(scriptHash)}
}

func (p *PolicyID) Bytes() []byte {
	return p.bs.Bytes()
}

func (p *PolicyID) String() string {
	return p.bs.String()
}

type AssetName struct {
	bs   cbor.ByteString
	name string
}

func NewAssetName(name string) AssetName {
	return AssetName{bs: cbor.NewByteString([]byte(name)), name: name}
}

func (an *AssetName) Bytes() []byte {
	return an.bs.Bytes()
}

func (an *AssetName) String() string {
	return an.name
}

type Assets struct {
	m map[cbor.ByteString]BigNum
}

func NewAssets() *Assets {
	return &Assets{m: make(map[cbor.ByteString]BigNum)}
}

func (a *Assets) Set(name AssetName, val BigNum) *Assets {
	a.m[name.bs] = val
	return a
}

func (a *Assets) Get(name AssetName) BigNum {
	return a.m[name.bs]
}

// MarshalCBOR implements cbor.Marshaler
func (a *Assets) MarshalCBOR() ([]byte, error) {
	return cborEnc.Marshal(a.m)
}

type MultiAsset struct {
	m map[cbor.ByteString]*Assets
}

func NewMultiAsset() *MultiAsset {
	return &MultiAsset{m: make(map[cbor.ByteString]*Assets)}
}

func (ma *MultiAsset) Set(policyID PolicyID, assets *Assets) *MultiAsset {
	ma.m[policyID.bs] = assets
	return ma
}

func (ma *MultiAsset) Get(policyID PolicyID) *Assets {
	return ma.m[policyID.bs]
}

func (ma *MultiAsset) NumPIDs() uint {
	return uint(len(ma.m))
}

func (ma *MultiAsset) NumAssets() uint {
	var num uint
	for _, assets := range ma.m {
		num += uint(len(assets.m))
	}
	return num
}

func (ma *MultiAsset) AssetsLength() uint {
	var sum uint
	for _, assets := range ma.m {
		for assetName := range assets.m {
			sum += uint(len(assetName.Bytes()))
		}
	}
	return sum
}

// MarshalCBOR implements cbor.Marshaler
func (ma *MultiAsset) MarshalCBOR() ([]byte, error) {
	return cborEnc.Marshal(ma.m)
}

type AddrKeyHash = Hash28

type PoolKeyHash = Hash28

type Hash28 []byte

// NewHash28 creates a new Hash28 from a hex encoded string.
func NewHash28(h string) (Hash28, error) {
	hash := make([]byte, 28)
	b, err := hex.DecodeString(h)
	if err != nil {
		return hash, err
	}
	copy(hash[:], b)
	return hash, nil
}

// String returns the hex encoding representation of a Hash28.
func (h Hash28) String() string {
	return hex.EncodeToString(h[:])
}

type Hash32 []byte

// NewHash32 creates a new Hash32 from a hex encoded string.
func NewHash32(h string) (Hash32, error) {
	hash := make([]byte, 32)
	b, err := hex.DecodeString(h)
	if err != nil {
		return hash, err
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

type UnitInterval = Rational

type Rational struct {
	_ struct{} `cbor:",toarray"`
	P uint64
	Q uint64
}

// MarshalCBOR implements cbor.Marshaler
func (r *Rational) MarshalCBOR() ([]byte, error) {
	type rational Rational

	// Register tag 30 for rational numbers
	tags, err := r.tagSet(rational{})
	if err != nil {
		return nil, err
	}

	em, err := cbor.CanonicalEncOptions().EncModeWithTags(tags)
	if err != nil {
		return nil, err
	}

	return em.Marshal(rational(*r))
}

// UnmarshalCBOR implements cbor.Unmarshaler
func (r *Rational) UnmarshalCBOR(data []byte) error {
	type rational Rational

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

func (r *Rational) tagSet(contentType interface{}) (cbor.TagSet, error) {
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
