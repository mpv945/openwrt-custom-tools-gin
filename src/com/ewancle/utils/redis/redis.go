package redis

import (
	"context"
	"errors"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/go-redsync/redsync/v4"
	goredis "github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/mpv945/openwrt-custom-tools-gin/src/com/ewancle/utils/json"

	"github.com/redis/go-redis/v9"
)

var Client *redis.Client
var Ctx = context.Background()

var Rs *redsync.Redsync

func Init(addr string, password string) {

	Client = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
	})
	/*rdb := redis.NewClient(&redis.Options{
		Addr:         "127.0.0.1:6379",
		Password:     "",
		DB:           0,

		PoolSize:     20, // PoolSize = CPU * 10
		MinIdleConns: 5,  // MinIdleConns >= CPU

		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,

		PoolTimeout: 4 * time.Second,

		MaxConnAge: 30 * time.Minute,
	})
	*/
	//_, err := Client.Ping(Ctx).Result()

	//ctx := context.Background()

	_, err := Client.Ping(Ctx).Result()
	if err != nil {
		log.Fatal("Redis connect failed:", err)
	}

	log.Println("Redis connected")
	// 初始化 Redsync
	pool := goredis.NewPool(Client)

	Rs = redsync.New(pool)
}

// String 操作

func SetString(key string, value interface{}, ttl time.Duration) error {
	// 设置 key 时直接指定 TTL：
	// err := rdb.Set(ctx, "user:1001", "tom", 10*time.Minute).Err()
	return Client.Set(Ctx, key, value, ttl).Err()
	//先 set 再 expire
	//rdb.Set(ctx, "user:1001", "tom", 0)
	//rdb.Expire(ctx, "user:1001", 10*time.Minute)
}

// 过期时间

func SetExpire(key string, ttl time.Duration) error {
	return Client.Expire(Ctx, key, ttl).Err()
}

func GetExpire(key string) (time.Duration, error) {
	return Client.TTL(Ctx, key).Result()
}

func DeleteExpire(key string) (bool, error) {
	return Client.Persist(Ctx, key).Result()
}

// RandomTTL 全部缓存 1 小时,随机秒 ：使用：rdb.Set(ctx, key, value, RandomTTL(time.Hour))
func RandomTTL(base time.Duration) time.Duration {
	return base + time.Duration(rand.Intn(300))*time.Second
}

func GetString(key string) (string, error) {
	return Client.Get(Ctx, key).Result()
}

func DeleteKey(key string) error {
	return Client.Del(Ctx, key).Err()
}

// Hash 操作

func HSet(key string, values map[string]interface{}) error {
	return Client.HSet(Ctx, key, values).Err()
}

func HGet(key string, field string) (string, error) {
	return Client.HGet(Ctx, key, field).Result()
}

func HGetAll(key string) (map[string]string, error) {
	return Client.HGetAll(Ctx, key).Result()
}

// List 操作

func LPush(key string, values ...interface{}) error {
	return Client.LPush(Ctx, key, values...).Err()
}

func RPop(key string) (string, error) {
	return Client.RPop(Ctx, key).Result()
}

func LRange(key string, start, stop int64) ([]string, error) {
	return Client.LRange(Ctx, key, start, stop).Result()
}

// Set 操作

func SAdd(key string, members ...interface{}) error {
	return Client.SAdd(Ctx, key, members...).Err()
}

func SMembers(key string) ([]string, error) {
	return Client.SMembers(Ctx, key).Result()
}

func SRem(key string, members ...interface{}) error {
	return Client.SRem(Ctx, key, members...).Err()
}

// Redis Stream（生产级消息队列）

const StreamName = "order_stream"

func ProduceMessage(values map[string]interface{}) (string, error) {

	id, err := Client.XAdd(Ctx, &redis.XAddArgs{
		Stream: StreamName,
		MaxLen: 100000,
		Approx: true,
		Values: values,
	}).Result()

	return id, err
}

const (
	GroupName    = "order_group"
	ConsumerName = "consumer_1"
)

func StartConsumer() {

	//rdb := redisclient.RDB

	// 创建消费者组
	err := Client.XGroupCreateMkStream(Ctx,
		StreamName,
		GroupName,
		//"0", //从历史第一条消息开始消费
		"$", //只消费创建之后的新消息
	).Err()

	if err != nil && !strings.Contains(err.Error(), "BUSYGROUP") {
		log.Fatal(err)
	}

	for {

		streams, err := Client.XReadGroup(Ctx, &redis.XReadGroupArgs{
			Group:    GroupName,
			Consumer: ConsumerName,
			Streams:  []string{StreamName, ">"},
			Count:    10,
			//Block:    0,
			Block: 5 * time.Second,
		}).Result()

		if err != nil {

			if errors.Is(err, redis.Nil) {
				log.Println("Block时间内 通道还没有没有数据")
				continue
			}

			log.Println("read error:", err)
			continue
		}

		for _, stream := range streams {
			for _, msg := range stream.Messages {
				log.Println("Stream receive:", msg.Values)
				// ack
				Client.XAck(Ctx, StreamName, GroupName, msg.ID)
			}
		}
	}
}

// 生产级锁封装

type DistributedLock struct {
	mutex *redsync.Mutex
}

func NewDistributedLock(key string) *DistributedLock {

	mutex := Rs.NewMutex(
		key,
		redsync.WithExpiry(10*time.Second), // 锁过期时间
		redsync.WithTries(3),               // 获取锁重试次数
		redsync.WithRetryDelay(500*time.Millisecond),
	)

	return &DistributedLock{
		mutex: mutex,
	}
}
func (l *DistributedLock) Lock() error {
	return l.mutex.LockContext(Ctx)
}

func (l *DistributedLock) Unlock() (bool, error) {
	return l.mutex.Unlock()
}

// Pub/Sub 消息实现

type Message struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

func Publish(channel string, msg Message) error {

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return Client.Publish(Ctx, channel, data).Err()
}

func Subscribe(channel string, handler func(Message)) {

	PubNub := Client.Subscribe(Ctx, channel)

	ch := PubNub.Channel()

	for msg := range ch {

		var m Message

		err := json.Unmarshal([]byte(msg.Payload), &m)
		if err != nil {
			log.Println("message decode error:", err)
			continue
		}

		handler(m)
	}
}
