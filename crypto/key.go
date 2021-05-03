package crypto

import (
	"crypto/sha256"
	"crypto/sha512"
	"fmt"

	"github.com/echovl/ed25519"
	"github.com/tyler-smith/go-bip39"
	"golang.org/x/crypto/pbkdf2"
)

type KeyPair struct {
	Priv      []byte
	Pub       []byte
	ChainCode []byte
}

// XPriv includes the 32 byte chain code
type XPriv []byte
type XPub []byte

func GenerateMasterKey(entropy []byte, password string) XPriv {
	xpriv := pbkdf2.Key([]byte(password), entropy, 4096, 96, sha512.New)

	xpriv[0] &= 0xf8
	xpriv[31] = (xpriv[31] & 0x1f) | 0x40

	return xpriv
}

func GenerateMnemonic(entropy []byte) (string, error) {
	if l := len(entropy); l != 20 {
		return "", fmt.Errorf("crypto: bad entropy size %d", l)
	}
	return bip39.NewMnemonic(entropy)
}

func NewKeyPair() (KeyPair, error) {
	kp := KeyPair{}

	for {
		pub, priv, err := ed25519.GenerateKey(nil)
		if err != nil {
			return KeyPair{}, nil
		}

		if !isValidPrivateKey(priv) {
			fmt.Println("Invalid extended private key")
			continue
		}

		hpriv := make([]byte, 33)
		hpriv[0] = 0x1
		copy(hpriv[1:], priv[:32])
		chainCode := sha256.Sum256(hpriv)

		kp.Pub = pub
		kp.Priv = priv[:32]
		kp.ChainCode = chainCode[:]

		return kp, nil
	}
}

func (kp *KeyPair) Sign(message []byte) []byte {
	priv := make([]byte, 64)

	copy(priv[:32], kp.Priv)
	copy(priv[32:], kp.Pub)

	return ed25519.Sign(ed25519.PrivateKey(priv), message)
}

func (kp *KeyPair) Verify(message, sig []byte) bool {
	return ed25519.Verify(kp.Pub, message, sig)
}

func isValidPrivateKey(priv []byte) bool {
	return (priv[31] & (1 << 5)) == 0
}
