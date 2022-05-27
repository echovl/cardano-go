package wallet

import (
	"encoding/json"
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
	skeys   []crypto.ExtendedSigningKey
	pkeys   []crypto.ExtendedVerificationKey
	rootKey crypto.ExtendedSigningKey
	node    node.Node
	network types.Network
}

// Transfer sends an amount of lovelace to the receiver address and returns the transaction hash
//TODO: remove hardcoded protocol parameters, these parameters must be obtained using the cardano node
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

	builder := tx.NewTxBuilder(types.ProtocolParams{
		MinimumUtxoValue: 1000000,
		MinFeeA:          44,
		MinFeeB:          155381,
	})

	keys := make(map[int]crypto.ExtendedSigningKey)
	for i, utxo := range pickedUtxos {
		for _, key := range w.skeys {
			vkey := key.ExtendedVerificationKey()
			addr := types.NewEnterpriseAddress(vkey, w.network)
			if addr == utxo.Spender {
				keys[i] = key
			}
		}
	}

	if len(keys) != len(pickedUtxos) {
		panic("not enough keys")
	}

	for i, utxo := range pickedUtxos {
		skey := keys[i]
		vkey := skey.ExtendedVerificationKey()
		builder.AddInput(vkey, utxo.TxHash, utxo.Index, utxo.Amount)
	}
	builder.AddOutput(receiver, amount)

	// Calculate and set ttl
	tip, err := w.node.Tip()
	if err != nil {
		return nil, err
	}
	builder.SetTtl(uint64(tip.Slot + 1200))

	changeAddress := pickedUtxos[0].Spender
	err = builder.AddFee(changeAddress)
	if err != nil {
		return nil, err
	}
	for _, key := range keys {
		builder.Sign(key)
	}
	tx := builder.Build()

	return w.node.SubmitTx(tx)
}

// Balance returns the total lovelace amount of the wallet.
func (w *Wallet) Balance() (types.Coin, error) {
	var balance types.Coin
	utxos, err := w.findUtxos()
	if err != nil {
		return 0, nil
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
	newKey := crypto.DeriveSigningKey(w.rootKey, index)
	w.skeys = append(w.skeys, newKey)
	return types.NewEnterpriseAddress(newKey.ExtendedVerificationKey(), w.network)
}

// Addresses returns all wallet's addresss.
func (w *Wallet) Addresses() []types.Address {
	addresses := make([]types.Address, len(w.skeys))
	for i, key := range w.skeys {
		addresses[i] = types.NewEnterpriseAddress(key.ExtendedVerificationKey(), w.network)
	}
	return addresses
}

func newWalletID() string {
	id, _ := gonanoid.Generate(walleIDAlphabet, 10)
	return "wallet_" + id
}

func newWallet(name, password string, entropy []byte) *Wallet {
	wallet := &Wallet{Name: name, ID: newWalletID()}
	rootKey := crypto.NewExtendedSigningKey(entropy, password)
	purposeKey := crypto.DeriveSigningKey(rootKey, purposeIndex)
	coinKey := crypto.DeriveSigningKey(purposeKey, coinTypeIndex)
	accountKey := crypto.DeriveSigningKey(coinKey, accountIndex)
	chainKey := crypto.DeriveSigningKey(accountKey, externalChainIndex)
	addr0Key := crypto.DeriveSigningKey(chainKey, 0)
	wallet.rootKey = chainKey
	wallet.skeys = []crypto.ExtendedSigningKey{addr0Key}
	return wallet
}

type walletDump struct {
	ID      string
	Name    string
	Keys    []crypto.ExtendedSigningKey
	RootKey crypto.ExtendedSigningKey
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
