package cardano

import (
	"encoding/hex"
	"fmt"
	"math/big"
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

func (v *Value) String() string {
	vMap := map[string]uint64{"lovelace": uint64(v.Coin)}
	for _, pool := range v.MultiAsset.Keys() {
		for _, assets := range v.MultiAsset.Get(pool).Keys() {
			fmt.Printf("%+v", assets)
			vMap[assets.String()] = uint64(v.MultiAsset.Get(pool).Get(assets))
		}
	}
	return fmt.Sprintf("%+v", vMap)
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
		reAssets := NewAssets()
		for assetName, value := range assets.m {
			reAssets.m[assetName] = value
		}
		result.MultiAsset.m[policy] = reAssets
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

// Compares and returns
//      -1 if v < rhs
//       0 if v == rhs
//       1 if v > rhs
//       2 if not comparable
func (v *Value) Cmp(rhs *Value) int {
	lrZero := v.Sub(rhs).IsZero()
	rlZero := rhs.Sub(v).IsZero()

	if !lrZero && !rlZero {
		return 2
	} else if lrZero && !rlZero {
		return -1
	} else if !lrZero && rlZero {
		return 1
	} else {
		return 0
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
	bs cbor.ByteString
}

func NewAssetName(name string) AssetName {
	return AssetName{bs: cbor.NewByteString([]byte(name))}
}

func (an *AssetName) Bytes() []byte {
	return an.bs.Bytes()
}

func (an AssetName) String() string {
	return string(an.bs.Bytes())
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

func (a *Assets) Keys() []AssetName {
	assetNames := []AssetName{}
	for k := range a.m {
		assetNames = append(assetNames, AssetName{bs: k})
	}
	return assetNames
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

func (ma *MultiAsset) Keys() []PolicyID {
	policyIDs := []PolicyID{}
	for id := range ma.m {
		policyIDs = append(policyIDs, PolicyID{bs: id})
	}
	return policyIDs
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

type MintAssets struct {
	m map[cbor.ByteString]*big.Int
}

func NewMintAssets() *MintAssets {
	return &MintAssets{m: make(map[cbor.ByteString]*big.Int)}
}

func (a *MintAssets) Set(name AssetName, val *big.Int) *MintAssets {
	a.m[name.bs] = val
	return a
}

func (a *MintAssets) Get(name AssetName) *big.Int {
	return a.m[name.bs]
}

func (a *MintAssets) Keys() []AssetName {
	assetNames := []AssetName{}
	for k := range a.m {
		assetNames = append(assetNames, AssetName{bs: k})
	}
	return assetNames
}

// MarshalCBOR implements cbor.Marshaler
func (a *MintAssets) MarshalCBOR() ([]byte, error) {
	return cborEnc.Marshal(a.m)
}

type Mint struct {
	m map[cbor.ByteString]*MintAssets
}

func NewMint() *Mint {
	return &Mint{m: make(map[cbor.ByteString]*MintAssets)}
}

func (m *Mint) Set(policyID PolicyID, assets *MintAssets) *Mint {
	m.m[policyID.bs] = assets
	return m
}

func (m *Mint) Get(policyID PolicyID) *MintAssets {
	return m.m[policyID.bs]
}

func (m *Mint) Keys() []PolicyID {
	policyIDs := []PolicyID{}
	for id := range m.m {
		policyIDs = append(policyIDs, PolicyID{bs: id})
	}
	return policyIDs
}

func (m *Mint) MultiAsset() *MultiAsset {
	ma := NewMultiAsset()
	for policy, mintAssets := range m.m {
		assets := NewAssets()
		for assetName, value := range mintAssets.m {
			posVal := value.Abs(value)
			if posVal.IsUint64() {
				assets.m[assetName] = BigNum(posVal.Uint64())
			} else {
				panic("MintAsset value cannot be represented as a uint64")
			}
		}
		ma.m[policy] = assets
	}
	return ma
}

func (ma *Mint) NumPIDs() uint {
	return uint(len(ma.m))
}

func (ma *Mint) NumAssets() uint {
	var num uint
	for _, assets := range ma.m {
		num += uint(len(assets.m))
	}
	return num
}

func (ma *Mint) AssetsLength() uint {
	var sum uint
	for _, assets := range ma.m {
		for assetName := range assets.m {
			sum += uint(len(assetName.Bytes()))
		}
	}
	return sum
}

// MarshalCBOR implements cbor.Marshaler
func (ma *Mint) MarshalCBOR() ([]byte, error) {
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
