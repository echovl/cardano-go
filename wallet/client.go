package wallet

import (
	"fmt"

	"github.com/echovl/cardano-go/crypto"
	"github.com/echovl/cardano-go/types"
	"github.com/tyler-smith/go-bip39"
)

// Client provides a clean interface for creating, saving and deleting Wallets.
type Client struct {
	opts    *Options
	network types.Network
}

// NewClient builds a new Client using cardano-cli as the default connection
// to the Blockhain.
//
// It uses BadgerDB as the default Wallet storage.
func NewClient(opts *Options) *Client {
	opts.init()
	cl := &Client{opts: opts, network: opts.Node.Network()}
	return cl
}

// Close closes all the resources used by the Client.
func (c *Client) Close() {
	c.opts.DB.Close()
}

// CreateWallet creates a new Wallet using a secure entropy and password,
// returning a Wallet with its corresponding 24 word mnemonic
func (c *Client) CreateWallet(name, password string) (*Wallet, string, error) {
	entropy := newEntropy(entropySizeInBits)
	mnemonic := crypto.NewMnemonic(entropy)
	wallet := newWallet(name, password, entropy)
	wallet.node = c.opts.Node
	wallet.network = c.network
	err := c.opts.DB.SaveWallet(wallet)
	if err != nil {
		return nil, "", err
	}
	return wallet, mnemonic, nil
}

// RestoreWallet restores a Wallet from a mnemonic and password.
func (c *Client) RestoreWallet(name, password, mnemonic string) (*Wallet, error) {
	entropy, err := bip39.EntropyFromMnemonic(mnemonic)
	if err != nil {
		return nil, err
	}
	wallet := newWallet(name, password, entropy)
	wallet.node = c.opts.Node
	wallet.network = c.network
	err = c.opts.DB.SaveWallet(wallet)
	if err != nil {
		return nil, err
	}

	return wallet, nil
}

// SaveWallet saves a Wallet in the Client's storage.
func (c *Client) SaveWallet(w *Wallet) error {
	return c.opts.DB.SaveWallet(w)
}

// Wallets returns the list of Wallets currently saved in the Client's storage.
func (c *Client) Wallets() ([]*Wallet, error) {
	wallets, err := c.opts.DB.GetWallets(c.network)
	if err != nil {
		return nil, err
	}
	for i := range wallets {
		wallets[i].node = c.opts.Node
	}
	return wallets, nil
}

// Wallet returns a Wallet with the given id from the Client's storage.
func (c *Client) Wallet(id string) (*Wallet, error) {
	wallets, err := c.Wallets()
	if err != nil {
		return nil, err
	}
	for _, w := range wallets {
		if w.ID == id {
			return w, nil
		}
	}
	return nil, fmt.Errorf("wallet %v not found", id)
}

// DeleteWallet removes a Wallet with the given id from the Client's storage.
func (c *Client) DeleteWallet(id string) error {
	return c.opts.DB.DeleteWallet(id)
}
