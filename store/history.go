package store

import (
	"encoding/json"
	"fmt"
	"github.com/mrChex/gustav/models"
	"go.etcd.io/bbolt"
	"time"
)

func (s *Store) HistoryPush(chatID int64, msg models.Message) error {
	msg.CreatedAt = time.Now()
	bucketName := []byte(fmt.Sprintf("chat-%d", chatID))
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal json: %w", err)
	}
	id := []byte(msg.CreatedAt.Format(time.RFC3339Nano))
	err = s.db.Update(func(tx *bbolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(bucketName)
		if err != nil {
			return fmt.Errorf("create bucket: %w", err)
		}
		return bucket.Put(id, data)
	})
	if err != nil {
		return fmt.Errorf("update: %w", err)
	}
	return nil
}

// GetHistory returns chat history from end to begin
func (s *Store) GetHistory(chatID int64, limit, offset int) ([]models.Message, error) {
	bucketName := []byte(fmt.Sprintf("chat-%d", chatID))
	var messages []models.Message
	err := s.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(bucketName)
		if bucket == nil {
			return nil
		}
		c := bucket.Cursor()
		i := 0
		for k, v := c.Last(); k != nil; k, v = c.Prev() {
			if i < offset {
				i++
				continue
			}
			m, err := models.NewMessageFromJSON(v)
			if err != nil {
				return fmt.Errorf("unmarshal json: %w", err)
			}
			messages = append(messages, m)
			if len(messages) >= limit {
				break
			}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("view: %w", err)
	}
	return messages, nil
}
