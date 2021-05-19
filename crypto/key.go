package crypto

import (
	"crypto/sha512"

	"github.com/echovl/ed25519"
	"github.com/tyler-smith/go-bip39"
	"golang.org/x/crypto/pbkdf2"
)

// XSigningKey is the extended private key appended with the chain code
type XSigningKey []byte

// XVerificationKey is the public key appended with the chain code
type XVerificationKey []byte

func GenerateMasterKey(entropy []byte, password string) XSigningKey {
	key := pbkdf2.Key([]byte(password), entropy, 4096, 96, sha512.New)

	key[0] &= 0xf8
	key[31] = (key[31] & 0x1f) | 0x40

	return key
}

func GenerateMnemonic(entropy []byte) string {
	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		panic(err)
	}
	return mnemonic
}

func (xsk *XSigningKey) XVerificationKey() XVerificationKey {
	xvk := make([]byte, 64)
	pk := ed25519.PublicKeyFrom(ed25519.ExtendedPrivateKey((*xsk)[:64]))
	cc := (*xsk)[64:]

	copy(xvk[:32], pk)
	copy(xvk[32:], cc)

	return xvk
}
