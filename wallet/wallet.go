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
	externalChainIndex uint32 = 0x0
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

// GenerateAddress generates a new payment address and adds it to the wallet.
func (w *Wallet) GenerateAddress(network NetworkType) Address {
	chain := w.ExternalChain
	index := uint32(len(chain.Childs))
	xsk := crypto.DeriveChildSigningKey(chain.Root.Xsk, index)
	w.ExternalChain.Childs = append(w.ExternalChain.Childs, keyPair{xsk, xsk.XVerificationKey()})

	if w.db != nil {
		w.db.SaveWallet(w)
	}

	return newEnterpriseAddress(xsk.XVerificationKey(), network)
}

// AddAddresses returns all wallet's addresss.
func (w *Wallet) Addresses(network NetworkType) []Address {
	addresses := make([]Address, len(w.ExternalChain.Childs))
	for i, kp := range w.ExternalChain.Childs {
		addresses[i] = newEnterpriseAddress(kp.Xvk, network)
	}

	return addresses
}

func NewWalletID() WalletID {
	id, _ := gonanoid.Generate(walleIDAlphabet, 10)
	return WalletID("wallet_" + id)
}

// AddWallet creates a brand new wallet using a secure entropy and password.
// This function will return a new wallet with its corresponding 24 word mnemonic
func AddWallet(name, password string, db WalletDB) (*Wallet, string, error) {
	entropy := newEntropy(entropySizeInBits)
	mnemonic := crypto.GenerateMnemonic(entropy)

	wallet := newWallet(entropy, password)
	wallet.Name = name
	err := db.SaveWallet(wallet)
	if err != nil {
		return nil, "", err
	}

	return wallet, mnemonic, nil
}

func RestoreWallet(name, mnemonic, password string, db WalletDB) (*Wallet, error) {
	entropy, err := bip39.EntropyFromMnemonic(mnemonic)
	if err != nil {
		return nil, err
	}

	wallet := newWallet(entropy, password)
	wallet.Name = name
	err = db.SaveWallet(wallet)
	if err != nil {
		return nil, err
	}

	return wallet, nil
}

func GetWallets(db WalletDB) ([]Wallet, error) {
	wallets, err := db.GetWallets()
	if err != nil {
		return nil, err
	}
	for i := range wallets {
		wallets[i].db = db
	}
	return wallets, nil
}

func GetWallet(id WalletID, db WalletDB) (*Wallet, error) {
	wallets, err := GetWallets(db)
	if err != nil {
		return nil, err
	}
	for _, w := range wallets {
		if w.ID == id {
			return &w, nil
		}
	}

	return nil, fmt.Errorf("wallet %v not found", id)
}

func DeleteWallet(id WalletID, db WalletDB) error {
	return db.DeleteWallet(id)
}

func newWallet(entropy []byte, password string) *Wallet {
	wallet := &Wallet{ID: NewWalletID()}

	rootXsk := crypto.GenerateMasterKey(entropy, password)
	purposeXsk := crypto.DeriveChildSigningKey(rootXsk, purposeIndex)
	coinXsk := crypto.DeriveChildSigningKey(purposeXsk, coinTypeIndex)
	accountXsk := crypto.DeriveChildSigningKey(coinXsk, accountIndex)
	chainXsk := crypto.DeriveChildSigningKey(accountXsk, externalChainIndex)

	addr0Xsk := crypto.DeriveChildSigningKey(chainXsk, 0)

	wallet.ExternalChain = keyPairChain{
		Root:   newKeyPairFromXsk(chainXsk),
		Childs: []keyPair{newKeyPairFromXsk(addr0Xsk)},
	}

	return wallet
}
