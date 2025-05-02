package ratelimiter

import (
	"encoding/json"
	"os"
	"sync"
)

type FileStorage struct {
	path string
	mu   sync.Mutex
}

func NewFileStorage(path string) *FileStorage {
	return &FileStorage{path: path}
}

func (s *FileStorage) Save(buckets map[string]*TokenBucket) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := json.Marshal(buckets)
	if err != nil {
		return err
	}

	return os.WriteFile(s.path, data, 0644)
}

func (s *FileStorage) Load() (map[string]*TokenBucket, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.path)
	if os.IsNotExist(err) {
		return make(map[string]*TokenBucket), nil
	}
	if err != nil {
		return nil, err
	}

	var buckets map[string]*TokenBucket
	if err := json.Unmarshal(data, &buckets); err != nil {
		return nil, err
	}

	for _, b := range buckets {
		b.mu = sync.Mutex{}
	}
	return buckets, nil
}
