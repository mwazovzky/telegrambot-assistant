package repository

// CacheClient defines the interface for key-value storage operations.
type CacheClient interface {
	Get(key string) (string, error)
	Set(key string, value string) error
}

// CacheRepository implements ResponseStore using a cache backend (Redis).
type CacheRepository struct {
	client CacheClient
}

func NewCachedRepository(client CacheClient) *CacheRepository {
	return &CacheRepository{client}
}

func (r *CacheRepository) GetResponseID(key string) (string, error) {
	return r.client.Get(key)
}

func (r *CacheRepository) SetResponseID(key string, responseID string) error {
	return r.client.Set(key, responseID)
}
