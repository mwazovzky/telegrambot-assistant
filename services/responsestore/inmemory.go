package responsestore

// InmemoryStore implements ResponseStore using in-memory storage.
type InmemoryStore struct {
	data map[string]string
}

func NewInmemoryStore() *InmemoryStore {
	return &InmemoryStore{data: make(map[string]string)}
}

func (r *InmemoryStore) GetResponseID(key string) (string, error) {
	return r.data[key], nil
}

func (r *InmemoryStore) SetResponseID(key string, responseID string) error {
	r.data[key] = responseID
	return nil
}
