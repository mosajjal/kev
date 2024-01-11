package server

// implements a KV interface based on BadgerDB v2
import (
	badger "github.com/dgraph-io/badger/v4"
)

type BadgerKV struct {
	DB *badger.DB
}

// NewBadgerKV returns a new BadgerKV based on the folder path provided
func NewBadgerKV(opts badger.Options) (KV, error) {
	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}
	db.DropAll()
	return &BadgerKV{DB: db}, nil
}

func (b *BadgerKV) Close() error {
	return b.DB.Close()
}

func (b *BadgerKV) Get(key string) (string, error) {
	v, err := b.DB.NewTransaction(false).Get([]byte(key))
	if err != nil {
		return "", err
	}
	return v.String(), nil
}
func (b *BadgerKV) Set(key string, value string) error {
	return b.DB.NewTransaction(true).Set([]byte(key), []byte(value))
}

func (b *BadgerKV) GetAllKeys() []string {
	keys := make([]string, 0)
	iterator := b.DB.NewTransaction(false).NewIterator(badger.DefaultIteratorOptions)
	for iterator.Rewind(); iterator.Valid(); iterator.Next() {
		keys = append(keys, string(iterator.Item().Key()))
	}
	return keys
}

func (b *BadgerKV) Delete(key string) error {
	return b.DB.NewTransaction(true).Delete([]byte(key))
}
