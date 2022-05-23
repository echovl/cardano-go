package cardano

import (
	"encoding/json"
	"fmt"
	"strconv"

	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/tclairet/cardano-go/crypto"
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
	node    cardanoNode
	network Network
}

func (w *Wallet) SetNetwork(net Network) {
	w.network = net
}

// Transfer sends an amount of lovelace to the receiver address
//TODO: remove hardcoded protocol parameters, these parameters must be obtained using the cardano node
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

	keys := make(map[int]crypto.ExtendedSigningKey)
	for i, utxo := range pickedUtxos {
		for _, key := range w.skeys {
			vkey := key.ExtendedVerificationKey()
			address := NewEnterpriseAddress(vkey, w.network)
			if address == utxo.Address {
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
		builder.AddInput(vkey, utxo.TxId, utxo.Index, utxo.Amount)
	}
	builder.AddOutput(receiver, amount)

	// Calculate and set ttl
	tip, err := w.node.QueryTip()
	if err != nil {
		return err
	}
	builder.SetTtl(tip.Slot + 1200)

	changeAddress := pickedUtxos[0].Address
	err = builder.AddFee(changeAddress)
	if err != nil {
		return err
	}
	for _, key := range keys {
		builder.Sign(key)
	}
	tx := builder.Build()
	return w.node.SubmitTx(tx)
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
		addrUtxos, err := w.node.QueryUtxos(addr)
		if err != nil {
			return nil, err
		}
		walletUtxos = append(walletUtxos, addrUtxos...)
	}
	return walletUtxos, nil
}

// AddAddress generates a new payment address and adds it to the wallet.
func (w *Wallet) AddAddress() Address {
	index := uint32(len(w.skeys))
	newKey := crypto.DeriveSigningKey(w.rootKey, index)
	w.skeys = append(w.skeys, newKey)
	return NewEnterpriseAddress(newKey.ExtendedVerificationKey(), w.network)
}

// Addresses returns all wallet's addresss.
func (w *Wallet) Addresses() []Address {
	addresses := make([]Address, len(w.skeys))
	for i, key := range w.skeys {
		addresses[i] = NewEnterpriseAddress(key.ExtendedVerificationKey(), w.network)
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
}

func (w *Wallet) marshal() ([]byte, error) {
	wd := &walletDump{
		ID:      w.ID,
		Name:    w.Name,
		Keys:    w.skeys,
		RootKey: w.rootKey,
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
	return nil
}

func ParseUint64(s string) (uint64, error) {
	parsed, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0, err
	}
	return parsed, nil
}

var newEntropy = func(bitSize int) []byte {
	entropy, _ := bip39.NewEntropy(bitSize)
	return entropy
}
