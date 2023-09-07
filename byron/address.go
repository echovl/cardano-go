package byron

import (
	"errors"
	"github.com/btcsuite/btcutil/base58"
	"github.com/echovl/cardano-go/crypto"
	"github.com/echovl/cardano-go/internal/cbor"
	"hash/crc32"
)

var (
	ErrInvalidByronAddress  = errors.New("invalid byron address")
	ErrInvalidByronChecksum = errors.New("invalid byron checksum")
)

type ByronAddressAttributes struct {
	Payload []byte `cbor:"1,keyasint,omitempty"`
	Network *uint8 `cbor:"2,keyasint,omitempty"`
}

type ByronAddress struct {
	Hash       []byte
	Attributes ByronAddressAttributes
	Tag        uint
}

const (
	XPubSize              = 64
	ByronPubKeyTag uint   = 0
	ByronTag       uint64 = 24
)

// Bytes returns byte slice represantation of the Address.
func (b *ByronAddress) Bytes() (bytes []byte) {
	bytes, _ = b.MarshalCBOR()
	return bytes
}

func (b *ByronAddress) Bech32() string {
	panic("not suport")
}

// String returns base58 encoded byron address.
func (b *ByronAddress) String() (str string) {
	return base58.Encode(b.Bytes())
}

// NetworkInfo returns NetworkInfo{ProtocolMagigic and NetworkId}.
func (b *ByronAddress) NetworkInfo() byte {
	if b.Attributes.Network == nil {
		return 1
	}
	return 0
}

// MarshalCBOR returns a cbor encoded byte slice of the base address.
func (b *ByronAddress) MarshalCBOR() (bytes []byte, err error) {
	raw, err := cbor.Marshal([]interface{}{b.Hash, b.Attributes, b.Tag})
	if err != nil {
		return nil, err
	}
	return cbor.Marshal([]interface{}{
		cbor.Tag{Number: 24, Content: raw},
		uint64(crc32.ChecksumIEEE(raw)),
	})
}

// UnmarshalCBOR deserializes raw byron address, encoded in cbor, into a Byron Address.
func (b *ByronAddress) UnmarshalCBOR(data []byte) error {
	type RawAddr struct {
		_        struct{} `cbor:",toarray"`
		Tag      cbor.Tag
		Checksum uint32
	}

	var rawAddr RawAddr

	if err := cbor.Unmarshal(data, &rawAddr); err != nil {
		return err
	}

	rawTag, ok := rawAddr.Tag.Content.([]byte)
	if !ok || rawAddr.Tag.Number != 24 {
		return ErrInvalidByronAddress
	}

	cheksum := crc32.ChecksumIEEE(rawTag)
	if rawAddr.Checksum != cheksum {
		return ErrInvalidByronChecksum
	}

	var byron struct {
		_      struct{} `cbor:",toarray"`
		Hashed []byte
		Attrs  ByronAddressAttributes
		Tag    uint
	}

	if err := cbor.Unmarshal(rawTag, &byron); err != nil {
		return err
	}

	if len(byron.Hashed) != 28 || byron.Tag != 0 {
		return errors.New("")
	}

	*b = ByronAddress{
		Hash:       byron.Hashed,
		Attributes: byron.Attrs,
		Tag:        byron.Tag,
	}

	return nil
}

// Pref returns the string prefix for the base address. "" for byron address since it has no prefix.
func (b *ByronAddress) Prefix() string {
	return ""
}

func NewSimpleLegacyAddress(xpub []byte, hdPath []byte) (*ByronAddress, error) {
	var attr ByronAddressAttributes
	attr.Payload, _ = cbor.Marshal(hdPath)
	hashed, err := newHashedSpending(xpub, attr)
	if err != nil {
		return nil, err
	}
	return &ByronAddress{Hash: hashed, Attributes: attr, Tag: ByronPubKeyTag}, nil
}

func newHashedSpending(xpub []byte, attr ByronAddressAttributes) ([]byte, error) {
	if len(xpub) != XPubSize {
		return nil, errors.New("xpub size should be 64 bytes")
	}
	spend, err := cbor.Marshal([]interface{}{ByronPubKeyTag, xpub})
	if err != nil {
		return nil, err
	}
	buf, err := cbor.Marshal([]interface{}{ByronPubKeyTag, cbor.RawMessage(spend), attr})
	if err != nil {
		return nil, err
	}
	return crypto.Sha3AndBlake2b224(buf)
}
