package crypto

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/tyler-smith/go-bip39"
)

const (
	mnemonic                   = "eight country switch draw meat scout mystery blade tip drift useless good keep usage title"
	masterKeyWithoutPassphrase = "c065afd2832cd8b087c4d9ab7011f481ee1e0721e78ea5dd609f3ab3f156d245d176bd8fd4ec60b4731c3918a2a72a0226c0cd119ec35b47e4d55884667f552a23f7fdcd4a10c6cd2c7393ac61d877873e248f417634aa3d812af327ffe9d620"
	masterKeyWithPassphrase    = "70531039904019351e1afb361cd1b312a4d0565d4ff9f8062d38acf4b15cce41d7b5738d9c893feea55512a3004acb0d222c35d3e3d5cde943a15a9824cbac59443cf67e589614076ba01e354b1a432e0e6db3b59e37fc56b5fb0222970a010e"
	passphrase                 = "foo"
)

func TestExtendedSigningKeyWithoutPassphrase(t *testing.T) {
	entropy, _ := bip39.EntropyFromMnemonic(mnemonic)

	got := NewExtendedSigningKey(entropy, "")
	want, _ := hex.DecodeString(masterKeyWithoutPassphrase)

	if bytes.Compare(got, want) != 0 {
		t.Errorf("invalid master key\ngot: %x\nwant: %x\n", got, want)
	}
}

func TestExtendedSigningKeyWithPassphrase(t *testing.T) {
	entropy, _ := bip39.EntropyFromMnemonic(mnemonic)

	got := NewExtendedSigningKey(entropy, passphrase)
	want, _ := hex.DecodeString(masterKeyWithPassphrase)

	if bytes.Compare(got, want) != 0 {
		t.Errorf("invalid master key\ngot: %x\nwant: %x\n", got, want)
	}
}
