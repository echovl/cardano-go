package wallet

import (
	"fmt"

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
	walleIDAlphabet           = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
)

var newEntropy = func(bitSize int) []byte {
	entropy, _ := bip39.NewEntropy(bitSize)
	return entropy
}

type WalletID string

type Wallet struct {
	ID            WalletID
	Name          string
	ExternalChain keyPairChain
	internalChain keyPairChain
	stakeChain    keyPairChain
	db            WalletDB
}

func (w *Wallet) Keys() []crypto.XVerificationKey {
	xvks := make([]crypto.XVerificationKey, len(w.ExternalChain.Childs))
	for i, kp := range w.ExternalChain.Childs {
		xvks[i] = kp.Xvk
	}
	return xvks
}

func (w *Wallet) NewAddress() (Address, error) {
	chain := w.ExternalChain
	index := uint32(len(chain.Childs))
	xsk := crypto.DeriveChildSigningKey(chain.Root.Xsk, index)
	w.ExternalChain.Childs = append(w.ExternalChain.Childs, keyPair{xsk, xsk.XVerificationKey()})

	if w.db != nil {
		w.db.SaveWallet(w)
	}

	return newEnterpriseAddress(xsk.XVerificationKey(), Testnet)
}

func (w *Wallet) Addresses() ([]Address, error) {
	addresses := make([]Address, len(w.ExternalChain.Childs))
	for i, kp := range w.ExternalChain.Childs {
		addr, err := newEnterpriseAddress(kp.Xvk, Testnet)
		if err != nil {
			return nil, err
		}
		addresses[i] = addr
	}

	return addresses, nil
}

func NewWalletID() WalletID {
	id, _ := gonanoid.Generate(walleIDAlphabet, 10)
	return WalletID("wl_" + id)
}

// AddWallet creates a brand new wallet using a secure entropy and password.
// This function will return a new wallet with its corresponding 24 word mnemonic
func AddWallet(name, password string, db WalletDB) (*Wallet, string, error) {
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

	db.SaveWallet(&wallet)

	return &wallet, mnemonic, nil
}

func RestoreWallet(mnemonic, password string, db WalletDB) (*Wallet, error) {
	wallet := Wallet{Name: "test", ID: NewWalletID()}

	entropy, err := bip39.EntropyFromMnemonic(mnemonic)
	if err != nil {
		return nil, err
	}

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

	db.SaveWallet(&wallet)

	return &wallet, nil
}

func GetWallets(db WalletDB) []Wallet {
	wallets := db.GetWallets()
	for i := range wallets {
		wallets[i].db = db
	}
	return wallets
}

func GetWallet(id WalletID, db WalletDB) (*Wallet, error) {
	wallets := GetWallets(db)
	for _, w := range wallets {
		if w.ID == id {
			return &w, nil
		}
	}
	return nil, fmt.Errorf("wallet %v not found", id)
}
