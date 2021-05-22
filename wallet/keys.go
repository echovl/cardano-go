package wallet

import "github.com/echovl/cardano-wallet/crypto"

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
