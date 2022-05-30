package crypto

import (
	"crypto/sha512"

	"github.com/echovl/bech32"
	"github.com/echovl/ed25519"
	"golang.org/x/crypto/pbkdf2"
)

// XPrvKey is the extended private key (64 bytes) appended with the chain code (32 bytes).
type XPrvKey []byte

// NewXPrvKey creates a new extended private key from a bech32 encoded private key.
func NewXPrvKey(bech string) (XPrvKey, error) {
	_, xsk, err := bech32.DecodeToBase256(bech)
	return xsk, err
}

func NewXPrvKeyFromEntropy(entropy []byte, password string) XPrvKey {
	key := pbkdf2.Key([]byte(password), entropy, 4096, 96, sha512.New)

	key[0] &= 0xf8
	key[31] = (key[31] & 0x1f) | 0x40

	return key
}

// Bech32 returns the private key encoded as bech32.
func (prv XPrvKey) Bech32(prefix string) string {
	bech, err := bech32.EncodeFromBase256(prefix, prv)
	if err != nil {
		panic(err)
	}
	return bech
}

// XPubKey returns the XPubKey derived from the extended private key.
func (prv XPrvKey) XPubKey() XPubKey {
	xvk := make([]byte, 64)
	pk := ed25519.PublicKeyFrom(ed25519.ExtendedPrivateKey(prv[:64]))
	cc := prv[64:]

	copy(xvk[:32], pk)
	copy(xvk[32:], cc)

	return xvk
}

// XPubKey returns the XPubKey derived from the extended private key.
func (prv XPrvKey) PubKey() PubKey {
	return PubKey(ed25519.PublicKeyFrom(ed25519.ExtendedPrivateKey(prv[:64])))
}

func (prv *XPrvKey) Sign(message []byte) []byte {
	pk := ed25519.ExtendedPrivateKey((*prv)[:64])
	return ed25519.SignExtended(pk, message)
}

// XPubKey is the public key (32 bytes) appended with the chain code (32 bytes).
type XPubKey []byte

// NewXPubKey creates a new extended public key from a bech32 encoded extended public key.
func NewXPubKey(bech string) (XPubKey, error) {
	_, xsk, err := bech32.DecodeToBase256(bech)
	return xsk, err
}

// XPubKey returns the PubKey from the extended public key.
func (pub XPubKey) PubKey() PubKey {
	return PubKey(pub[:32])
}

// NewPubKey creates a new public key from a bech32 encoded public key.
func NewPubKey(bech string) (PubKey, error) {
	_, xsk, err := bech32.DecodeToBase256(bech)
	return xsk, err
}

// Verify reports whether sig is a valid signature of message by the extended public key.
func (pub XPubKey) Verify(message, sig []byte) bool {
	return pub.PubKey().Verify(message, sig)
}

// PubKey is a edd25519 public key.
type PubKey []byte

// Verify reports whether sig is a valid signature of message by the public key.
func (pub PubKey) Verify(message, signature []byte) bool {
	return ed25519.Verify(ed25519.PublicKey(pub), message, signature)
}
