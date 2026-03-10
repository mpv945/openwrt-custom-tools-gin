package redis

import (
	"github.com/redis/go-redis/v9"
)

type UVService struct {
	client *redis.Client
}

func NewUVService(client *redis.Client) *UVService {
	return &UVService{client: client}
}

func (u *UVService) AddUV(key string, userID string) error {

	return u.client.PFAdd(Ctx, key, userID).Err()
}

func (u *UVService) CountUV(key string) (int64, error) {

	return u.client.PFCount(Ctx, key).Result()
}
