package wallet

import (
	"fmt"

	"github.com/echovl/bech32"
	"github.com/echovl/cardano-wallet/crypto"
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
	xsk crypto.XSigningKey
	xvk crypto.XVerificationKey
}

type keyPairChain struct {
	root   keyPair
	childs []keyPair
}

type Wallet struct {
	ID            string
	Name          string
	externalChain keyPairChain
	internalChain keyPairChain
	stakeChain    keyPairChain
}

// AddWallet creates a brand new wallet using a secure entropy and password.
// This function will return a new wallet with its corresponding 24 word mnemonic
func AddWallet(password string) (*Wallet, string, error) {
	wallet := Wallet{Name: "test", ID: "wallet1"}

	entropy := newEntropy(entropySizeInBits)
	mnemonic, err := crypto.GenerateMnemonic(entropy)
	if err != nil {
		return nil, "", err
	}

	rootXsk := crypto.GenerateMasterKey(entropy, password)
	fmt.Println(bech32.EncodeFromBase256("root_xsk", rootXsk))
	purposeXsk := crypto.DeriveChildSigningKey(rootXsk, purposeIndex)
	coinXsk := crypto.DeriveChildSigningKey(purposeXsk, coinTypeIndex)
	accountXsk := crypto.DeriveChildSigningKey(coinXsk, accountIndex)
	chainXsk := crypto.DeriveChildSigningKey(accountXsk, externalChainIndex)

	addr0Xsk := crypto.DeriveChildSigningKey(chainXsk, 0)

	wallet.externalChain = keyPairChain{
		root:   keyPair{chainXsk, chainXsk.XVerificationKey()},
		childs: []keyPair{{addr0Xsk, addr0Xsk.XVerificationKey()}},
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

	wallet.externalChain = keyPairChain{
		root:   keyPair{chainXsk, chainXsk.XVerificationKey()},
		childs: []keyPair{{addr0Xsk, addr0Xsk.XVerificationKey()}},
	}

	return &wallet, nil
}
