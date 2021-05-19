package db

import (
	"bytes"
	"encoding/json"

	"github.com/dgraph-io/badger/v3"
	"github.com/echovl/cardano-wallet/wallet"
)

type BadgerDB struct {
	db *badger.DB
}

func NewBadgerDB() *BadgerDB {
	db, err := badger.Open(badger.DefaultOptions("/tmp/badger").WithLoggingLevel(badger.ERROR))
	if err != nil {
		panic(err)
	}

	return &BadgerDB{db}
}

func (bdb *BadgerDB) Close() {
	bdb.db.Close()
}

func (bdb *BadgerDB) SaveWallet(w *wallet.Wallet) error {
	walletBuffer := &bytes.Buffer{}
	json.NewEncoder(walletBuffer).Encode(w)

	err := bdb.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(w.ID), walletBuffer.Bytes())
	})
	if err != nil {
		return err
	}

	return nil
}

func (bdb *BadgerDB) GetWallets() []wallet.Wallet {
	wallets := []wallet.Wallet{}
	bdb.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			value, err := item.ValueCopy(nil)
			if err != nil {
				return err
			}

			wallet := wallet.Wallet{}
			json.NewDecoder(bytes.NewBuffer(value)).Decode(&wallet)
			wallets = append(wallets, wallet)
		}
		return nil
	})
	return wallets
}

func (bdb *BadgerDB) DeleteWallet(id wallet.WalletID) error {
	err := bdb.db.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(id))
	})

	return err
}
