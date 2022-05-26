package wallet

import (
	"github.com/dgraph-io/badger/v3"
)

type DB interface {
	SaveWallet(*Wallet) error
	GetWallets() ([]*Wallet, error)
	DeleteWallet(string) error
	Close()
}

type badgerDB struct {
	db *badger.DB
}

func newBadgerDB() *badgerDB {
	db, err := badger.Open(badger.DefaultOptions("/tmp/badger").WithLoggingLevel(badger.ERROR))
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

func (bdb *badgerDB) GetWallets() ([]*Wallet, error) {
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
			wallets = append(wallets, wallet)
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
