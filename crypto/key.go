package crypto

import (
	"crypto/sha512"

	"github.com/echovl/bech32"
	"github.com/echovl/ed25519"
	"github.com/tyler-smith/go-bip39"
	"golang.org/x/crypto/pbkdf2"
)

// XPrv is the extended private key (64 bytes) appended with the chain code (32 bytes).
type XPrv []byte

func NewXPrv(entropy []byte, password string) XPrv {
	key := pbkdf2.Key([]byte(password), entropy, 4096, 96, sha512.New)

	key[0] &= 0xf8
	key[31] = (key[31] & 0x1f) | 0x40

	return key
}

// NewXPrvFromBech32 creates a new XPrv from a bech32 encoded private key.
func NewXPrvFromBech32(bech string) (XPrv, error) {
	_, xsk, err := bech32.DecodeToBase256(bech)
	return xsk, err
}

// String returns the private key encoded as string.
func (xsk XPrv) String() string {
	return xsk.Bech32()
}

// Bech32 returns the private key encoded as bech32.
func (xsk XPrv) Bech32() string {
	bech, err := bech32.EncodeFromBase256("xprv", xsk)
	if err != nil {
		panic(err)
	}
	return bech
}

// PublicKey returns the XPub derived from the XPrv.
func (xsk *XPrv) PublicKey() XPub {
	xvk := make([]byte, 64)
	pk := ed25519.PublicKeyFrom(ed25519.ExtendedPrivateKey((*xsk)[:64]))
	cc := (*xsk)[64:]

	copy(xvk[:32], pk)
	copy(xvk[32:], cc)

	return xvk
}

func (xsk *XPrv) Sign(message []byte) []byte {
	pk := ed25519.ExtendedPrivateKey((*xsk)[:64])
	return ed25519.SignExtended(pk, message)
}

// XPub is the public key (32 bytes) appended with the chain code (32 bytes).
type XPub []byte

func (xvk *XPub) Verify(message, signature []byte) bool {
	pk := ed25519.PublicKey((*xvk)[:32])
	return ed25519.Verify(pk, message, signature)
}

func NewMnemonic(entropy []byte) string {
	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		panic(err)
	}

	return mnemonic
}
