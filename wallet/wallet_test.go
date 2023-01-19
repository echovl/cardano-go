package wallet

import (
	"testing"

	"github.com/echovl/cardano-go"
	"github.com/echovl/cardano-go/internal/bech32"
	"github.com/tyler-smith/go-bip39"
)

func TestGenerateAddress(t *testing.T) {
	for _, testVector := range testVectors {
		client := NewClient(&Options{})
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

		paymentAddr1, err := w.AddAddress()
		if err != nil {
			t.Fatal(err)
		}

		addrXsk1 := bech32From("addr_xsk", w.addrKeys[1])
		addrXvk1 := bech32From("addr_xvk", w.addrKeys[1].XPubKey())

		if addrXsk1 != testVector.addrXsk1 {
			t.Errorf("invalid addrXsk1 :\ngot: %v\nwant: %v", addrXsk1, testVector.addrXsk1)
		}

		if addrXvk1 != testVector.addrXvk1 {
			t.Errorf("invalid addrXvk1 :\ngot: %v\nwant: %v", addrXvk1, testVector.addrXvk1)
		}

		if paymentAddr1.Bech32() != testVector.paymentEnterpriseAddr1 {
			t.Errorf("invalid paymentEnterpriseAddr1:\ngot: %v\nwant: %v", paymentAddr1, testVector.paymentEnterpriseAddr1)
		}
	}
}

type MockNode struct {
	utxos []cardano.UTxO
}

func (n *MockNode) UTxOs(addr cardano.Address) ([]cardano.UTxO, error) {
	return n.utxos, nil
}

func (n *MockNode) Tip() (*cardano.NodeTip, error) {
	return &cardano.NodeTip{}, nil
}

func (n *MockNode) SubmitTx(tx *cardano.Tx) (*cardano.Hash32, error) {
	return nil, nil
}

func (n *MockNode) ProtocolParams() (*cardano.ProtocolParams, error) {
	return &cardano.ProtocolParams{}, nil
}

func (n *MockNode) Network() cardano.Network {
	return cardano.Testnet
}

func TestWalletBalance(t *testing.T) {
	client := NewClient(&Options{
		Node: &MockNode{utxos: []cardano.UTxO{{Amount: cardano.NewValue(100)}, {Amount: cardano.NewValue(33)}}},
	})
	w, _, err := client.CreateWallet("test", "")
	if err != nil {
		t.Error(err)
	}

	got, err := w.Balance()
	if err != nil {
		t.Error(err)
	}
	want := cardano.NewValue(133)

	if got.Cmp(want) != 0 {
		t.Errorf("invalid balance:\ngot: %v\nwant: %v", got, want)
	}
}

func bech32From(hrp string, bytes []byte) string {
	enc, _ := bech32.EncodeFromBase256(hrp, bytes)
	return enc
}
