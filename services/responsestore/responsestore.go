package responsestore

// ResponseStore stores and retrieves OpenAI response IDs for conversation continuity.
type ResponseStore interface {
	GetResponseID(key string) (string, error)
	SetResponseID(key string, responseID string) error
}
