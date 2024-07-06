package store

import (
	"fmt"
	"go.etcd.io/bbolt"
	"strings"
)

func (s *Store) MemoryUpsert(chatID int64, category, key, value string) error {
	bucketName := []byte(fmt.Sprintf("memory-%d-%s", chatID, category))

	err := s.db.Update(func(tx *bbolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(bucketName)
		if err != nil {
			return fmt.Errorf("create bucket: %w", err)
		}
		return bucket.Put([]byte(key), []byte(value))
	})

	if err != nil {
		return fmt.Errorf("update: %w", err)
	}

	return nil
}

func (s *Store) MemoryAllCategories(chatID int64) ([]string, error) {
	var categories []string
	err := s.db.View(func(tx *bbolt.Tx) error {
		return tx.ForEach(func(name []byte, _ *bbolt.Bucket) error {
			if len(name) > 8 && strings.HasPrefix(string(name), fmt.Sprintf("memory-%d", chatID)) {
				categories = append(categories, string(name[8:]))
			}
			return nil
		})
	})

	if err != nil {
		return nil, fmt.Errorf("view: %w", err)
	}

	return categories, nil
}

func (s *Store) MemoryAllCategoryKeys(chatID int64, category string) ([]string, error) {
	bucketName := []byte(fmt.Sprintf("memory-%d-%s", chatID, category))
	var keys []string
	err := s.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(bucketName)
		if bucket == nil {
			return nil
		}
		return bucket.ForEach(func(k, _ []byte) error {
			keys = append(keys, string(k))
			return nil
		})
	})

	if err != nil {
		return nil, fmt.Errorf("view: %w", err)
	}

	return keys, nil
}

func (s *Store) MemoryMapValues(chatID int64, category string, keys []string) ([]string, error) {
	bucketName := []byte(fmt.Sprintf("memory-%d-%s", chatID, category))
	values := make([]string, len(keys))
	err := s.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(bucketName)
		if bucket == nil {
			return nil
		}
		for keyI, key := range keys {
			value := bucket.Get([]byte(key))
			if value != nil {
				values[keyI] = string(value)
			}
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("view: %w", err)
	}

	return values, nil
}

func (s *Store) MemoryDelete(chatID int64, category, key string) error {
	bucketName := []byte(fmt.Sprintf("memory-%d-%s", chatID, category))

	err := s.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(bucketName)
		if bucket == nil {
			return nil
		}
		return bucket.Delete([]byte(key))
	})

	if err != nil {
		return fmt.Errorf("update: %w", err)
	}

	return nil
}
