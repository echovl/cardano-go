package wallet

import (
	"fmt"

	"github.com/echovl/bech32"
	"github.com/echovl/cardano-wallet/crypto"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/tyler-smith/go-bip39"
)

const (
	entropySizeInBits         = 160
	purposeIndex       uint32 = 1852 + 0x80000000
	coinTypeIndex      uint32 = 1815 + 0x80000000
	accountIndex       uint32 = 0x80000000
	externalChainIndex uint32 = 0x80000000
)

type keyPair struct {
	Xsk crypto.XSigningKey
	Xvk crypto.XVerificationKey
}

type keyPairChain struct {
	Root   keyPair
	Childs []keyPair
}

type WalletID string

type Wallet struct {
	ID            WalletID
	Name          string
	ExternalChain keyPairChain
	internalChain keyPairChain
	stakeChain    keyPairChain
}

func (w *Wallet) Keys() []crypto.XVerificationKey {
	xvks := make([]crypto.XVerificationKey, len(w.ExternalChain.Childs))
	for i, kp := range w.ExternalChain.Childs {
		xvks[i] = kp.Xvk
	}
	return xvks
}

func NewWalletID() WalletID {
	id, _ := gonanoid.Generate("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz", 10)
	return WalletID("wl_" + id)
}

// AddWallet creates a brand new wallet using a secure entropy and password.
// This function will return a new wallet with its corresponding 24 word mnemonic
func AddWallet(name, password string) (*Wallet, string, error) {
	wallet := Wallet{Name: name, ID: NewWalletID()}
	entropy := newEntropy(entropySizeInBits)
	mnemonic := crypto.GenerateMnemonic(entropy)

	rootXsk := crypto.GenerateMasterKey(entropy, password)
	purposeXsk := crypto.DeriveChildSigningKey(rootXsk, purposeIndex)
	coinXsk := crypto.DeriveChildSigningKey(purposeXsk, coinTypeIndex)
	accountXsk := crypto.DeriveChildSigningKey(coinXsk, accountIndex)
	chainXsk := crypto.DeriveChildSigningKey(accountXsk, externalChainIndex)

	addr0Xsk := crypto.DeriveChildSigningKey(chainXsk, 0)

	wallet.ExternalChain = keyPairChain{
		Root:   keyPair{chainXsk, chainXsk.XVerificationKey()},
		Childs: []keyPair{{addr0Xsk, addr0Xsk.XVerificationKey()}},
	}

	return &wallet, mnemonic, nil
}

var newEntropy = func(bitSize int) []byte {
	entropy, _ := bip39.NewEntropy(bitSize)
	return entropy
}

func RestoreWallet(mnemonic, password string) (*Wallet, error) {
	wallet := Wallet{Name: "test", ID: "wallet1"}

	entropy, err := bip39.EntropyFromMnemonic(mnemonic)
	if err != nil {
		return nil, err
	}

	rootXsk := crypto.GenerateMasterKey(entropy, password)
	fmt.Println(bech32.EncodeFromBase256("root_xsk", rootXsk))
	purposeXsk := crypto.DeriveChildSigningKey(rootXsk, purposeIndex)
	coinXsk := crypto.DeriveChildSigningKey(purposeXsk, coinTypeIndex)
	accountXsk := crypto.DeriveChildSigningKey(coinXsk, accountIndex)
	chainXsk := crypto.DeriveChildSigningKey(accountXsk, externalChainIndex)

	addr0Xsk := crypto.DeriveChildSigningKey(chainXsk, 0)

	wallet.ExternalChain = keyPairChain{
		Root:   keyPair{chainXsk, chainXsk.XVerificationKey()},
		Childs: []keyPair{{addr0Xsk, addr0Xsk.XVerificationKey()}},
	}

	return &wallet, nil
}
