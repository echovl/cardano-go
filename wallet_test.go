package cardano

import (
	"testing"

	"github.com/echovl/bech32"
	"github.com/tyler-smith/go-bip39"
)

func TestGenerateAddress(t *testing.T) {
	for _, testVector := range testVectors {
		client := NewClient(WithDB(&MockDB{}))
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
		w.SetNetwork(Testnet)

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
	utxos []Utxo
}

func (prov *MockNode) QueryUtxos(addr Address) ([]Utxo, error) {
	return prov.utxos, nil
}

func (prov *MockNode) QueryTip() (NodeTip, error) {
	return NodeTip{}, nil
}

func (prov *MockNode) SubmitTx(tx Transaction) error {
	return nil
}

func TestWalletBalance(t *testing.T) {
	client := NewClient(WithDB(&MockDB{}))
	client.node = &MockNode{utxos: []Utxo{{Amount: 100}, {Amount: 33}}}
	w, _, err := client.CreateWallet("test", "")
	if err != nil {
		t.Error(err)
	}

	got, err := w.Balance()
	if err != nil {
		t.Error(err)
	}
	want := uint64(133)

	if got != want {
		t.Errorf("invalid balance :\ngot: %v\nwant: %v", got, want)
	}
}

func bech32From(hrp string, bytes []byte) string {
	enc, _ := bech32.EncodeFromBase256(hrp, bytes)
	return enc
}
