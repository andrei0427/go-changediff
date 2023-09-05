package services

import (
	"time"

	"github.com/patrickmn/go-cache"
)

type CacheService struct {
	cache *cache.Cache
}

func NewCacheService() *CacheService {
	return &CacheService{cache: cache.New(1*time.Hour, 2*time.Hour)}
}

func (s *CacheService) Set(key string, value any, duration *time.Duration) {
	var expiration = cache.DefaultExpiration
	if duration != nil {
		expiration = *duration
	}

	s.cache.Set(key, value, expiration)
}

func (s *CacheService) Get(key string) (interface{}, bool) {
	return s.cache.Get(key)
}
