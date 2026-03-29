package responsestore

// CacheClient defines the interface for key-value storage operations.
type CacheClient interface {
	Get(key string) (string, error)
	Set(key string, value string) error
}

// RedisStore implements ResponseStore using a Redis cache backend.
type RedisStore struct {
	client CacheClient
}

func NewRedisStore(client CacheClient) *RedisStore {
	return &RedisStore{client}
}

func (r *RedisStore) GetResponseID(key string) (string, error) {
	return r.client.Get(key)
}

func (r *RedisStore) SetResponseID(key string, responseID string) error {
	return r.client.Set(key, responseID)
}
