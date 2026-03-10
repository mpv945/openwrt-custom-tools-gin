package redis

import (
	"github.com/redis/go-redis/v9"
)

type PVService struct {
	client *redis.Client
}

func NewPVService(client *redis.Client) *PVService {
	return &PVService{client: client}
}

func (p *PVService) IncrPV(key string) (int64, error) {

	return p.client.Incr(Ctx, key).Result()
}

func (p *PVService) GetPV(key string) (int64, error) {

	return p.client.Get(Ctx, key).Int64()
}
