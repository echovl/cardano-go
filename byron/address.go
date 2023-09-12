package byron

import (
	"errors"
	"fmt"
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
	Payload []byte  `cbor:"1,keyasint,omitempty"`
	Network *uint32 `cbor:"2,keyasint,omitempty"`
}

type attributeCborType struct {
	HDPayload []byte `cbor:"1,keyasint,omitempty"`
	Network   []byte `cbor:"2,keyasint,omitempty"`
}

type ByronAddress struct {
	Hash       []byte
	Attributes ByronAddressAttributes
	Tag        uint64
}

const (
	XPubSize              = 64
	ByronPubKeyTag uint64 = 0
	ByronTag       uint64 = 24
	Hash28Size            = 28
)

// Bytes returns byte slice represantation of the Address.
func (b *ByronAddress) Bytes() (bytes []byte) {
	bytes, _ = b.MarshalCBOR()
	return bytes
}

func (b *ByronAddress) Bech32() string {
	return b.String()
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

func (a *ByronAddress) MarshalCBOR() ([]byte, error) {
	if len(a.Hash) != Hash28Size {
		return nil, errors.New("Invalid hash28 data")
	}
	raw, err := cbor.Marshal([]interface{}{a.Hash, a.Attributes, a.Tag})
	if err != nil {
		return nil, err
	}
	//fmt.Println("hex.encode:", hex.EncodeToString(raw))
	return cbor.Marshal([]interface{}{
		cbor.Tag{Number: ByronTag, Content: raw},
		uint64(crc32.ChecksumIEEE(raw)),
	})
}

func (a *ByronAddress) UnmarshalCBOR(data []byte) error {
	type RawAddress struct {
		_        struct{} `cbor:",toarray"`
		Tag      cbor.Tag
		Checksum uint64
	}

	var rawAddress RawAddress
	if err := cbor.Unmarshal(data, &rawAddress); err != nil {
		return fmt.Errorf("mashal raw: %s", err)
	}

	rawTag, ok := rawAddress.Tag.Content.([]byte)
	if !ok || rawAddress.Tag.Number != ByronTag {
		return errors.New("not a valid byron address")
	}

	checksum := crc32.ChecksumIEEE(rawTag)
	if rawAddress.Checksum != uint64(checksum) {
		return errors.New("checksum unmatched")
	}
	var got struct {
		_      struct{} `cbor:",toarray"`
		Hashed []byte
		Attrs  ByronAddressAttributes
		Tag    uint64
	}
	if err := cbor.Unmarshal(rawTag, &got); err != nil {
		return err
	}
	if len(got.Hashed) != Hash28Size || got.Tag != ByronPubKeyTag {
		return errors.New("Invalid byron hashed or type")
	}
	if a == nil {
		return errors.New("unmarshal to nil value")
	}
	*a = ByronAddress{Hash: got.Hashed, Attributes: got.Attrs, Tag: got.Tag}
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

func (a ByronAddressAttributes) MarshalCBOR() ([]byte, error) {
	var t = &attributeCborType{HDPayload: a.Payload}
	if a.Network != nil {
		network, err := cbor.Marshal(a.Network)
		if err != nil {
			return nil, err
		}
		t.Network = network
	}
	return cbor.Marshal(t)
}

func (a *ByronAddressAttributes) UnmarshalCBOR(data []byte) error {
	var t attributeCborType
	if err := cbor.Unmarshal(data, &t); err != nil {
		return err
	}

	var network *uint32
	if len(t.Network) != 0 {
		if err := cbor.Unmarshal(t.Network, &network); err != nil {
			return err
		}
	}

	*a = ByronAddressAttributes{
		Payload: t.HDPayload,
		Network: network,
	}
	return nil
}
