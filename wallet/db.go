package wallet

import (
	"os"
	"path"

	"github.com/dgraph-io/badger/v3"
	"github.com/echovl/cardano-go"
)

type DB interface {
	SaveWallet(*Wallet) error
	GetWallets(cardano.Network) ([]*Wallet, error)
	DeleteWallet(string) error
	Close()
}

type badgerDB struct {
	db *badger.DB
}

func newBadgerDB() *badgerDB {
	home, _ := os.UserHomeDir()
	db, err := badger.Open(badger.DefaultOptions(path.Join(home, ".cwallet/db")).WithLoggingLevel(badger.ERROR))
	if err != nil {
		panic(err)
	}
	return &badgerDB{db}
}

func (bdb *badgerDB) Close() {
	bdb.db.Close()
}

func (bdb *badgerDB) SaveWallet(w *Wallet) error {
	bytes, err := w.marshal()
	if err != nil {
		return err
	}
	err = bdb.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(w.ID), bytes)
	})
	if err != nil {
		return err
	}
	return nil
}

func (bdb *badgerDB) GetWallets(network cardano.Network) ([]*Wallet, error) {
	wallets := []*Wallet{}
	err := bdb.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			value, err := item.ValueCopy(nil)
			if err != nil {
				return err
			}
			wallet := &Wallet{}
			wallet.unmarshal(value)
			if wallet.network == network {
				wallets = append(wallets, wallet)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return wallets, nil
}

func (bdb *badgerDB) DeleteWallet(id string) error {
	err := bdb.db.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(id))
	})
	return err
}
