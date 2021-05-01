package crypto

import (
	"crypto/ed25519"
	"crypto/sha256"
	"fmt"
)

type KeyPair struct {
	Priv      []byte
	Pub       []byte
	ChainCode []byte
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

func (kp *KeyPair) Derive(index uint32) (KeyPair, error) {
	ckp := KeyPair{}

	deriveChildPrivateKey(*kp, index)

	return ckp, nil
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
