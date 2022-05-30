package wallet

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/echovl/cardano-go/crypto"
	"github.com/echovl/cardano-go/node"
	"github.com/echovl/cardano-go/tx"
	"github.com/echovl/cardano-go/types"
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

type Wallet struct {
	ID      string
	Name    string
	skeys   []crypto.XPrvKey
	pkeys   []crypto.XPubKey
	rootKey crypto.XPrvKey
	node    node.Node
	network types.Network
}

// Transfer sends an amount of lovelace to the receiver address and returns the transaction hash
func (w *Wallet) Transfer(receiver types.Address, amount types.Coin) (*types.Hash32, error) {
	// Calculate if the account has enough balance
	balance, err := w.Balance()
	if err != nil {
		return nil, err
	}
	if amount > balance {
		return nil, fmt.Errorf("Not enough balance, %v > %v", amount, balance)
	}

	// Find utxos that cover the amount to transfer
	pickedUtxos := []tx.UTxO{}
	utxos, err := w.findUtxos()
	pickedUtxosAmount := types.Coin(0)
	for _, utxo := range utxos {
		if pickedUtxosAmount > amount {
			break
		}
		pickedUtxos = append(pickedUtxos, utxo)
		pickedUtxosAmount += utxo.Amount
	}

	pparams, err := w.node.ProtocolParams()
	if err != nil {
		return nil, err
	}

	builder := tx.NewTxBuilder(pparams)

	keys := make(map[int]crypto.XPrvKey)
	for i, utxo := range pickedUtxos {
		for _, key := range w.skeys {
			payment, err := types.NewAddrKeyCredential(key.PubKey())
			if err != nil {
				return nil, err
			}
			addr, err := types.NewEnterpriseAddress(w.network, payment)
			if err != nil {
				return nil, err
			}
			if addr.Bech32() == utxo.Spender.Bech32() {
				keys[i] = key
			}
		}
	}

	if len(keys) != len(pickedUtxos) {
		return nil, errors.New("not enough keys")
	}

	for _, utxo := range pickedUtxos {
		builder.AddInputs(&tx.TxInput{TxHash: utxo.TxHash, Index: utxo.Index, Amount: utxo.Amount})
	}
	builder.AddOutputs(&tx.TxOutput{Address: receiver, Amount: amount})

	tip, err := w.node.Tip()
	if err != nil {
		return nil, err
	}
	builder.SetTTL(tip.Slot + 1200)
	for _, key := range keys {
		builder.Sign(key.Bech32("addr_xsk"))
	}
	changeAddress := pickedUtxos[0].Spender
	if err = builder.AddChangeIfNeeded(changeAddress.Bech32()); err != nil {
		return nil, err
	}

	tx, err := builder.Build()
	if err != nil {
		return nil, err
	}

	return w.node.SubmitTx(tx)
}

// Balance returns the total lovelace amount of the wallet.
func (w *Wallet) Balance() (types.Coin, error) {
	var balance types.Coin
	utxos, err := w.findUtxos()
	if err != nil {
		return 0, err
	}
	for _, utxo := range utxos {
		balance += utxo.Amount
	}
	return balance, nil
}

func (w *Wallet) findUtxos() ([]tx.UTxO, error) {
	addrs, err := w.Addresses()
	if err != nil {
		return nil, err
	}
	walletUtxos := []tx.UTxO{}
	for _, addr := range addrs {
		addrUtxos, err := w.node.UTxOs(addr)
		if err != nil {
			return nil, err
		}
		walletUtxos = append(walletUtxos, addrUtxos...)
	}
	return walletUtxos, nil
}

// AddAddress generates a new payment address and adds it to the wallet.
func (w *Wallet) AddAddress() (types.Address, error) {
	index := uint32(len(w.skeys))
	newKey := w.rootKey.Derive(index)
	w.skeys = append(w.skeys, newKey)
	payment, err := types.NewAddrKeyCredential(newKey.PubKey())
	if err != nil {
		return types.Address{}, err
	}
	return types.NewEnterpriseAddress(w.network, payment)
}

// Addresses returns all wallet's addresss.
func (w *Wallet) Addresses() ([]types.Address, error) {
	addresses := make([]types.Address, len(w.skeys))
	for i, key := range w.skeys {
		payment, err := types.NewAddrKeyCredential(key.PubKey())
		if err != nil {
			return nil, err
		}
		addr, err := types.NewEnterpriseAddress(w.network, payment)
		if err != nil {
			return nil, err
		}
		addresses[i] = addr
	}
	return addresses, nil
}

func newWalletID() string {
	id, _ := gonanoid.Generate(walleIDAlphabet, 10)
	return "wallet_" + id
}

func newWallet(name, password string, entropy []byte) *Wallet {
	wallet := &Wallet{Name: name, ID: newWalletID()}
	rootKey := crypto.NewXPrvKeyFromEntropy(entropy, password)
	chainKey := rootKey.Derive(purposeIndex).
		Derive(coinTypeIndex).
		Derive(accountIndex).
		Derive(externalChainIndex)
	addr0Key := chainKey.Derive(0)
	wallet.rootKey = chainKey
	wallet.skeys = []crypto.XPrvKey{addr0Key}
	return wallet
}

type walletDump struct {
	ID      string
	Name    string
	Keys    []crypto.XPrvKey
	RootKey crypto.XPrvKey
	Network types.Network
}

func (w *Wallet) marshal() ([]byte, error) {
	wd := &walletDump{
		ID:      w.ID,
		Name:    w.Name,
		Keys:    w.skeys,
		RootKey: w.rootKey,
		Network: w.network,
	}
	bytes, err := json.Marshal(wd)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

func (w *Wallet) unmarshal(bytes []byte) error {
	wd := &walletDump{}
	err := json.Unmarshal(bytes, wd)
	if err != nil {
		return err
	}
	w.ID = wd.ID
	w.Name = wd.Name
	w.skeys = wd.Keys
	w.rootKey = wd.RootKey
	w.network = wd.Network
	return nil
}

var newEntropy = func(bitSize int) []byte {
	entropy, _ := bip39.NewEntropy(bitSize)
	return entropy
}
