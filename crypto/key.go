package crypto

import (
	"crypto/sha512"

	"github.com/echovl/ed25519"
	"github.com/tyler-smith/go-bip39"
	"golang.org/x/crypto/pbkdf2"
)

// ExtendedSigningKey is the extended private key (64 bytes) appended with the chain code (32 bytes)
type ExtendedSigningKey []byte

func (xsk *ExtendedSigningKey) Sign(message []byte) []byte {
	pk := ed25519.ExtendedPrivateKey((*xsk)[:64])
	return ed25519.SignExtended(pk, message)
}

// ExtendedVerificationKey is the public key (32 bytes) appended with the chain code (32 bytes)
type ExtendedVerificationKey []byte

func (xvk *ExtendedVerificationKey) Verify(message, signature []byte) bool {
	pk := ed25519.PublicKey((*xvk)[:32])
	return ed25519.Verify(pk, message, signature)
}

func NewExtendedSigningKey(entropy []byte, password string) ExtendedSigningKey {
	key := pbkdf2.Key([]byte(password), entropy, 4096, 96, sha512.New)

	key[0] &= 0xf8
	key[31] = (key[31] & 0x1f) | 0x40

	return key
}

func NewMnemonic(entropy []byte) string {
	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		panic(err)
	}
	return mnemonic
}

func (xsk *ExtendedSigningKey) ExtendedVerificationKey() ExtendedVerificationKey {
	xvk := make([]byte, 64)
	pk := ed25519.PublicKeyFrom(ed25519.ExtendedPrivateKey((*xsk)[:64]))
	cc := (*xsk)[64:]

	copy(xvk[:32], pk)
	copy(xvk[32:], cc)

	return xvk
}
