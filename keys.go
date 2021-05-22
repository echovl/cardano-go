package cardano

import "github.com/echovl/cardano-go/crypto"

type KeyPair struct {
	Xsk crypto.XSigningKey
	Xvk crypto.XVerificationKey
}

type KeyPairChain struct {
	Root   KeyPair
	Childs []KeyPair
}

func newKeyPairFromXsk(xsk crypto.XSigningKey) KeyPair {
	return KeyPair{xsk, xsk.XVerificationKey()}
}
