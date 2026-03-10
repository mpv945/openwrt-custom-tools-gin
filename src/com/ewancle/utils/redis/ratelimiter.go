package redis

import (
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type SlidingWindowLimiter struct {
	client *redis.Client
	limit  int
	window time.Duration
}

// 1 删除窗口外数据 2 统计当前窗口数量 3 判断是否超过限制 4 写入当前请求

func NewSlidingWindowLimiter(client *redis.Client, limit int, window time.Duration) *SlidingWindowLimiter {

	return &SlidingWindowLimiter{
		client: client,
		limit:  limit,
		window: window,
	}
}

func (l *SlidingWindowLimiter) Allow(key string) (bool, error) {

	now := time.Now().UnixMilli()
	start := now - l.window.Milliseconds()

	pipe := l.client.TxPipeline()

	pipe.ZRemRangeByScore(Ctx, key, "0", strconv.FormatInt(start, 10))

	count := pipe.ZCard(Ctx, key)

	pipe.ZAdd(Ctx, key, redis.Z{
		Score:  float64(now),
		Member: now,
	})

	pipe.Expire(Ctx, key, l.window)

	_, err := pipe.Exec(Ctx)
	if err != nil {
		return false, err
	}

	if count.Val() >= int64(l.limit) {
		return false, nil
	}

	return true, nil
}
