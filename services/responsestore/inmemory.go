package responsestore

import "fmt"

// InmemoryStore implements ResponseStore using in-memory storage.
type InmemoryStore struct {
	data map[string]string
}

func NewInmemoryStore() *InmemoryStore {
	return &InmemoryStore{data: make(map[string]string)}
}

func (r *InmemoryStore) GetResponseID(key string) (string, error) {
	value, ok := r.data[key]
	if !ok {
		return "", fmt.Errorf("key [%s] does not exist", key)
	}
	return value, nil
}

func (r *InmemoryStore) SetResponseID(key string, responseID string) error {
	r.data[key] = responseID
	return nil
}
