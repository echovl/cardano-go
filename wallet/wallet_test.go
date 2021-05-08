package wallet

import "testing"

const (
	wallet1Password         = "cardano"
	wallet1Mnemonic         = "pulse bomb sad rib provide impulse farm situate fix sea museum camp another life mountain"
	wallet1Address0 Address = "addr1v9fm439vlcjdt6qac3dusatryyj6nnzy7erjqycpcczmrqcxtya36"
	wallet1Address1 Address = "addr1vy6je4u34akzmta256y5svfmqlqfaahr37qpwtprzyu968g48s5x7"
)

var wallet1Entropy = []byte{173, 131, 38, 246, 220, 122, 206, 228, 84, 206, 77, 88, 24, 62, 70, 144, 96, 153, 2, 164}

type MockDB struct {
	calls int
}

func (db *MockDB) SaveWallet(w *Wallet) error {
	db.calls++
	return nil
}

func (db *MockDB) Close() {
	db.calls++
}

func (db *MockDB) GetWallets() []Wallet {
	db.calls++
	return []Wallet{}
}

func TestAddWallet(t *testing.T) {
	db := &MockDB{}
	newEntropy = func(bitSize int) []byte {
		return wallet1Entropy
	}

	w, mnemonic, err := AddWallet("test", wallet1Password, db)
	if err != nil {
		t.Error(err)
	}

	addresses, err := w.Addresses(Mainnet)
	if err != nil {
		t.Error(err)
	}

	if mnemonic != wallet1Mnemonic {
		t.Errorf("invalid mnemonic:\ngot: %v\nwant: %v", mnemonic, wallet1Mnemonic)
	}

	if addresses[0] != wallet1Address0 {
		t.Errorf("invalid address:\ngot: %v\nwant: %v", addresses[0], wallet1Address0)
	}

	if db.calls != 1 {
		t.Errorf("invalid address:\ngot: %v\nwant: %v", db.calls, 1)
	}
}

func TestRestoreWallet(t *testing.T) {
	db := &MockDB{}
	newEntropy = func(bitSize int) []byte {
		return wallet1Entropy
	}

	w, err := RestoreWallet(wallet1Mnemonic, wallet1Password, db)
	if err != nil {
		t.Error(err)
	}

	addresses, err := w.Addresses(Mainnet)
	if err != nil {
		t.Error(err)
	}

	if addresses[0] != wallet1Address0 {
		t.Errorf("invalid address:\ngot: %v\nwant: %v", addresses[0], wallet1Address0)
	}

	if db.calls != 1 {
		t.Errorf("invalid db calls:\ngot: %v\nwant: %v", db.calls, 1)
	}
}

func TestGenerateAddress(t *testing.T) {
	db := &MockDB{}
	newEntropy = func(bitSize int) []byte {
		return wallet1Entropy
	}

	w, _, err := AddWallet("test", wallet1Password, db)
	if err != nil {
		t.Error(err)
	}

	addr1, err := w.GenerateAddress(Mainnet)

	if addr1 != wallet1Address1 {
		t.Errorf("invalid address:\ngot: %v\nwant: %v", addr1, wallet1Address1)
	}

	if db.calls != 1 {
		t.Errorf("invalid db calls:\ngot: %v\nwant: %v", db.calls, 1)
	}
}
