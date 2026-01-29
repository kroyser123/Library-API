package cache

import "time"

type Cache interface {
	Set(key string, value interface{}, ttl time.Duration) error
	Get(key string, value interface{}) error
	Delete(key string) error
	Clear() error
	Close() error
}
