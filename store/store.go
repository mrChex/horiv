package store

import (
	"go.etcd.io/bbolt"
	"log"
)

type Store struct {
	db *bbolt.DB
}

func NewStore(dbPath string) *Store {
	db, err := bbolt.Open(dbPath, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	return &Store{
		db: db,
	}
}

func (s *Store) Close() {
	if err := s.db.Close(); err != nil {
		log.Fatal(err)
	}
}
