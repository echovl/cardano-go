package wallet

import (
	"fmt"
	"strconv"

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
	ExternalChain KeyPairChain
	internalChain KeyPairChain
	stakeChain    KeyPairChain
	db            WalletDB
	provider      Provider
	network       Network
}

func (w *Wallet) SetProvider(p Provider) {
	w.provider = p
}

func (w *Wallet) SetNetwork(net Network) {
	w.network = net
}

func (w *Wallet) Transfer(receiver Address, amount uint64) error {
	// Calculate if the account has enough balance
	balance, err := w.Balance()
	if err != nil {
		return err
	}

	if amount > balance {
		return fmt.Errorf("Not enough balance, %v > %v", amount, balance)
	}

	// Find utxos that cover the amount to transfer
	pickedUtxos := []Utxo{}
	utxos, err := w.findUtxos()
	pickedUtxosAmount := uint64(0)
	for _, utxo := range utxos {
		if pickedUtxosAmount > amount {
			break
		}
		pickedUtxos = append(pickedUtxos, utxo)
		pickedUtxosAmount += utxo.Amount
	}

	builder := NewTxBuilder(ProtocolParams{
		MinimumUtxoValue: 1000000,
		MinFeeA:          44,
		MinFeeB:          155381,
	})

	keys := make(map[int]KeyPair)
	for i, utxo := range pickedUtxos {
		for _, keyPair := range w.ExternalChain.Childs {
			keyPairAddress := newEnterpriseAddress(keyPair.Xvk, w.network)
			if keyPairAddress == utxo.Address {
				keys[i] = keyPair
			}
		}
	}

	if len(keys) != len(pickedUtxos) {
		panic("not enough keys")
	}

	for i, utxo := range pickedUtxos {
		xvk := keys[i].Xvk
		builder.AddInput(xvk, utxo.TxId, utxo.Index, utxo.Amount)
	}
	builder.AddOutput(receiver, amount)

	// Calculate and set ttl
	tip, err := w.provider.QueryTip()
	if err != nil {
		return err
	}
	builder.SetTtl(tip.Slot + 1200)

	changeAddress := pickedUtxos[0].Address
	err = builder.AddFee(changeAddress)
	if err != nil {
		return err
	}

	for _, keyPair := range keys {
		builder.Sign(keyPair.Xsk)
	}

	tx := builder.Build()

	fmt.Println(pretty(tx))

	err = w.provider.SubmitTx(tx)

	return err
}

// Balance returns the total lovelace amount of the wallet.
func (w *Wallet) Balance() (uint64, error) {
	var balance uint64
	utxos, err := w.findUtxos()
	if err != nil {
		return 0, nil
	}

	for _, utxo := range utxos {
		balance += utxo.Amount
	}

	return balance, nil
}

func (w *Wallet) findUtxos() ([]Utxo, error) {
	addresses := w.Addresses()
	walletUtxos := []Utxo{}
	for _, addr := range addresses {
		addrUtxos, err := w.provider.QueryUtxos(addr)
		if err != nil {
			return nil, err
		}
		walletUtxos = append(walletUtxos, addrUtxos...)

	}
	return walletUtxos, nil
}

// GenerateAddress generates a new payment address and adds it to the wallet.
func (w *Wallet) GenerateAddress() Address {
	chain := w.ExternalChain
	index := uint32(len(chain.Childs))
	xsk := crypto.DeriveChildSigningKey(chain.Root.Xsk, index)
	w.ExternalChain.Childs = append(w.ExternalChain.Childs, KeyPair{xsk, xsk.XVerificationKey()})

	if w.db != nil {
		w.db.SaveWallet(w)
	}

	return newEnterpriseAddress(xsk.XVerificationKey(), w.network)
}

// AddAddresses returns all wallet's addresss.
func (w *Wallet) Addresses() []Address {
	addresses := make([]Address, len(w.ExternalChain.Childs))
	for i, kp := range w.ExternalChain.Childs {
		addresses[i] = newEnterpriseAddress(kp.Xvk, w.network)
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

// TODO: Find utxos with amount to restore the number of addresses
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

	wallet.ExternalChain = KeyPairChain{
		Root:   newKeyPairFromXsk(chainXsk),
		Childs: []KeyPair{newKeyPairFromXsk(addr0Xsk)},
	}

	return wallet
}

func ParseUint64(s string) (uint64, error) {
	parsed, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0, err
	}
	return parsed, nil
}
