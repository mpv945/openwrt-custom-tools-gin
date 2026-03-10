package redis

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type Leaderboard struct {
	client *redis.Client
	key    string
}

func NewLeaderboard(client *redis.Client, key string) *Leaderboard {

	return &Leaderboard{
		client: client,
		key:    key,
	}
}

/*
| 功能     | Redis命令         |
| ------ | --------------- |
| 增加分数   | ZINCRBY         |
| 获取TOPN | ZREVRANGE       |
| 查询排名   | ZREVRANK        |
| 查询分数   | ZSCORE          |
| 清理尾部   | ZREMRANGEBYRANK |
*/

func (l *Leaderboard) AddScore(user string, score float64) error {

	return l.client.ZIncrBy(Ctx, l.key, score, user).Err()
}

func (l *Leaderboard) TopN(n int64) ([]redis.Z, error) {

	return l.client.ZRevRangeWithScores(Ctx, l.key, 0, n-1).Result()
}

func (l *Leaderboard) Rank(user string) (int64, error) {

	return l.client.ZRevRank(Ctx, l.key, user).Result()
}

func (l *Leaderboard) Score(user string) (float64, error) {

	return l.client.ZScore(Ctx, l.key, user).Result()
}

// 清理排行榜，只保留前 topN 名

func (l *Leaderboard) Trim(topN int64) error {

	_, err := l.client.ZRemRangeByRank(
		Ctx,
		l.key,
		topN,
		-1,
	).Result()

	return err
}

// 生产优化版本（Pipeline） 排行榜

func (l *Leaderboard) AddScoreAndTrim(
	ctx context.Context,
	user string,
	score float64,
	topN int64,
) error {

	pipe := l.client.Pipeline()

	pipe.ZIncrBy(ctx, l.key, score, user)

	pipe.ZRemRangeByRank(ctx, l.key, topN, -1)

	_, err := pipe.Exec(ctx)

	return err
}
