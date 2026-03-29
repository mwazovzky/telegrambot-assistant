package repository

import "fmt"

// InmemoryRepository implements ResponseStore using in-memory storage.
type InmemoryRepository struct {
	data map[string]string
}

func NewInmemoryRepository() *InmemoryRepository {
	return &InmemoryRepository{data: make(map[string]string)}
}

func (r *InmemoryRepository) GetResponseID(key string) (string, error) {
	value, ok := r.data[key]
	if !ok {
		return "", fmt.Errorf("key [%s] does not exist", key)
	}
	return value, nil
}

func (r *InmemoryRepository) SetResponseID(key string, responseID string) error {
	r.data[key] = responseID
	return nil
}
