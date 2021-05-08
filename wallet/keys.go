package wallet

import "github.com/echovl/cardano-wallet/crypto"

type keyPair struct {
	Xsk crypto.XSigningKey
	Xvk crypto.XVerificationKey
}

type keyPairChain struct {
	Root   keyPair
	Childs []keyPair
}

func newKeyPairFromXsk(xsk crypto.XSigningKey) keyPair {
	return keyPair{xsk, xsk.XVerificationKey()}
}
