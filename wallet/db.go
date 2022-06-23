package wallet

import (
	"errors"
	"os"
	"path"

	"github.com/dgraph-io/badger/v3"
)

type DB interface {
	Put(*Wallet) error
	Get() ([]*Wallet, error)
	Delete(string) error
	Close() error
}

var (
	_ DB = (*badgerDB)(nil)
	_ DB = (*memoryDB)(nil)
)

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

func (bdb *badgerDB) Close() error {
	return bdb.db.Close()
}

func (bdb *badgerDB) Put(w *Wallet) error {
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

func (bdb *badgerDB) Get() ([]*Wallet, error) {
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

func (bdb *badgerDB) Delete(id string) error {
	err := bdb.db.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(id))
	})
	return err
}

type memoryDB struct {
	wm map[string]*Wallet
}

func newMemoryDB() DB {
	return &memoryDB{wm: make(map[string]*Wallet)}
}

func (db *memoryDB) Close() error {
	return nil
}

func (db *memoryDB) Put(w *Wallet) error {
	if w.ID == "" {
		return errors.New("empty wallet id")
	}
	db.wm[w.ID] = w
	return nil
}

func (db *memoryDB) Get() ([]*Wallet, error) {
	wallets := make([]*Wallet, len(db.wm))
	idx := 0
	for _, w := range db.wm {
		wallets[idx] = w
		idx++
	}
	return wallets, nil
}

func (db *memoryDB) Delete(id string) error {
	delete(db.wm, id)
	return nil
}
