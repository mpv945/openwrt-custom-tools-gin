package redis

import (
	"hash/fnv"

	"github.com/redis/go-redis/v9"
)

type BloomFilter struct {
	client *redis.Client
	key    string
	size   uint
}

func NewBloomFilter(client *redis.Client, key string, size uint) *BloomFilter {

	return &BloomFilter{
		client: client,
		key:    key,
		size:   size,
	}
}

func hash(data string, seed uint32) uint {

	h := fnv.New32a()
	_, err := h.Write([]byte(data))
	if err != nil {
		return 0
	}

	return uint(h.Sum32()+seed) % 1000000
}

func (b *BloomFilter) Add(value string) error {

	for i := 0; i < 3; i++ {

		offset := hash(value, uint32(i))

		err := b.client.SetBit(Ctx, b.key, int64(offset), 1).Err()
		if err != nil {
			return err
		}
	}

	return nil
}

func (b *BloomFilter) Exists(value string) (bool, error) {

	for i := 0; i < 3; i++ {

		offset := hash(value, uint32(i))

		bit, err := b.client.GetBit(Ctx, b.key, int64(offset)).Result()
		if err != nil {
			return false, err
		}

		if bit == 0 {
			return false, nil
		}
	}

	return true, nil
}
