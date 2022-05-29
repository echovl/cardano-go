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
	skeys   []crypto.XPrv
	pkeys   []crypto.XPub
	rootKey crypto.XPrv
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
	pickedUtxos := []tx.UTXO{}
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

	keys := make(map[int]crypto.XPrv)
	for i, utxo := range pickedUtxos {
		for _, key := range w.skeys {
			vkey := key.PublicKey()
			addr := types.NewEnterpriseAddress(vkey, w.network)
			if addr.String() == utxo.Spender.String() {
				keys[i] = key
			}
		}
	}

	if len(keys) != len(pickedUtxos) {
		return nil, errors.New("not enough keys")
	}

	for _, utxo := range pickedUtxos {
		builder.AddInputs(tx.TransactionInput{TxHash: utxo.TxHash, Index: utxo.Index, Amount: utxo.Amount})
	}
	builder.AddOutputs(tx.TransactionOutput{Address: receiver, Amount: amount})

	// Calculate and set ttl
	tip, err := w.node.Tip()
	if err != nil {
		return nil, err
	}
	builder.SetTTL(tip.Slot + 1200)

	for _, key := range keys {
		builder.Sign(key)
	}
	changeAddress := pickedUtxos[0].Spender
	if err = builder.AddFee(changeAddress); err != nil {
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

func (w *Wallet) findUtxos() ([]tx.UTXO, error) {
	addrs := w.Addresses()
	walletUtxos := []tx.UTXO{}
	for _, addr := range addrs {
		addrUtxos, err := w.node.UTXOs(addr)
		if err != nil {
			return nil, err
		}
		walletUtxos = append(walletUtxos, addrUtxos...)
	}
	return walletUtxos, nil
}

// AddAddress generates a new payment address and adds it to the wallet.
func (w *Wallet) AddAddress() types.Address {
	index := uint32(len(w.skeys))
	newKey := w.rootKey.DeriveXPrv(index)
	w.skeys = append(w.skeys, newKey)
	return types.NewEnterpriseAddress(newKey.PublicKey(), w.network)
}

// Addresses returns all wallet's addresss.
func (w *Wallet) Addresses() []types.Address {
	addresses := make([]types.Address, len(w.skeys))
	for i, key := range w.skeys {
		addresses[i] = types.NewEnterpriseAddress(key.PublicKey(), w.network)
	}
	return addresses
}

func newWalletID() string {
	id, _ := gonanoid.Generate(walleIDAlphabet, 10)
	return "wallet_" + id
}

func newWallet(name, password string, entropy []byte) *Wallet {
	wallet := &Wallet{Name: name, ID: newWalletID()}
	rootKey := crypto.NewXPrv(entropy, password)
	chainKey := rootKey.DeriveXPrv(purposeIndex).
		DeriveXPrv(coinTypeIndex).
		DeriveXPrv(accountIndex).
		DeriveXPrv(externalChainIndex)
	addr0Key := chainKey.DeriveXPrv(0)
	wallet.rootKey = chainKey
	wallet.skeys = []crypto.XPrv{addr0Key}
	return wallet
}

type walletDump struct {
	ID      string
	Name    string
	Keys    []crypto.XPrv
	RootKey crypto.XPrv
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
