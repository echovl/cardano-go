package wallet

import (
	"testing"

	"github.com/echovl/bech32"
	"github.com/echovl/cardano-go/node"
	"github.com/echovl/cardano-go/tx"
	"github.com/echovl/cardano-go/types"
	"github.com/tyler-smith/go-bip39"
)

func TestGenerateAddress(t *testing.T) {
	for _, testVector := range testVectors {
		client := NewClient(&Options{DB: &MockDB{}})
		defer client.Close()

		newEntropy = func(bitSize int) []byte {
			entropy, err := bip39.EntropyFromMnemonic(testVector.mnemonic)
			if err != nil {
				t.Error(err)
			}
			return entropy
		}

		w, _, err := client.CreateWallet("test", "")
		if err != nil {
			t.Error(err)
		}

		paymentAddr1 := w.AddAddress()

		addrXsk1 := bech32From("addr_xsk", w.skeys[1])
		addrXvk1 := bech32From("addr_xvk", w.skeys[1].ExtendedVerificationKey())

		if addrXsk1 != testVector.addrXsk1 {
			t.Errorf("invalid addrXsk1 :\ngot: %v\nwant: %v", addrXsk1, testVector.addrXsk1)
		}

		if addrXvk1 != testVector.addrXvk1 {
			t.Errorf("invalid addrXvk1 :\ngot: %v\nwant: %v", addrXvk1, testVector.addrXvk1)
		}

		if paymentAddr1 != testVector.paymentAddr1 {
			t.Errorf("invalid paymentAddr1:\ngot: %v\nwant: %v", paymentAddr1, testVector.paymentAddr1)
		}
	}
}

type MockNode struct {
	utxos []tx.UTXO
}

func (n *MockNode) UTXOs(addr types.Address) ([]tx.UTXO, error) {
	return n.utxos, nil
}

func (n *MockNode) Tip() (*node.NodeTip, error) {
	return &node.NodeTip{}, nil
}

func (n *MockNode) SubmitTx(tx tx.Transaction) (*types.Hash32, error) {
	return nil, nil
}

func (n *MockNode) Network() types.Network {
	return types.Testnet
}

func TestWalletBalance(t *testing.T) {
	client := NewClient(&Options{
		DB:   &MockDB{},
		Node: &MockNode{utxos: []tx.UTXO{{Amount: 100}, {Amount: 33}}},
	})
	w, _, err := client.CreateWallet("test", "")
	if err != nil {
		t.Error(err)
	}

	got, err := w.Balance()
	if err != nil {
		t.Error(err)
	}
	want := types.Coin(133)

	if got != want {
		t.Errorf("invalid balance :\ngot: %v\nwant: %v", got, want)
	}
}

func bech32From(hrp string, bytes []byte) string {
	enc, _ := bech32.EncodeFromBase256(hrp, bytes)
	return enc
}
