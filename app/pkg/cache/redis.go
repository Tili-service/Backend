package cache

import (
	"bytes"
	"context"
	"encoding/gob"
	"os"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

const DefaultTTL = 10 * time.Minute

func NewRedisClient(url string) *redis.Client {
	addr := strings.TrimSpace(url)
	if addr == "" {
		addr = strings.TrimSpace(os.Getenv("REDIS_URL"))
	}
	if addr == "" {
		addr = strings.TrimSpace(os.Getenv("REDIS_ADDR"))
	}
	if addr == "" {
		addr = "localhost:6379"
	}

	return redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     os.Getenv("REDIS_PASSWORD"),
		DB:           0,
		DialTimeout:  300 * time.Millisecond,
		ReadTimeout:  300 * time.Millisecond,
		WriteTimeout: 300 * time.Millisecond,
		PoolTimeout:  500 * time.Millisecond,
		MaxRetries:   1,
	})
}

func Get(ctx context.Context, client *redis.Client, key string, dest any) (bool, error) {
	if client == nil {
		return false, nil
	}

	data, err := client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	if err := gob.NewDecoder(bytes.NewReader(data)).Decode(dest); err != nil {
		return false, err
	}

	return true, nil
}

func Set(ctx context.Context, client *redis.Client, key string, value any, ttl time.Duration) error {
	if client == nil {
		return nil
	}

	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(value); err != nil {
		return err
	}

	if ttl <= 0 {
		ttl = DefaultTTL
	}

	return client.Set(ctx, key, buf.Bytes(), ttl).Err()
}

func DeletePrefix(ctx context.Context, client *redis.Client, prefix string) error {
	if client == nil {
		return nil
	}

	iter := client.Scan(ctx, 0, prefix+"*", 0).Iterator()
	const batchSize = 200
	keys := make([]string, 0, batchSize)
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
		if len(keys) >= batchSize {
			if err := client.Del(ctx, keys...).Err(); err != nil {
				return err
			}
			keys = keys[:0]
		}
	}
	if err := iter.Err(); err != nil {
		return err
	}
	if len(keys) > 0 {
		if err := client.Del(ctx, keys...).Err(); err != nil {
			return err
		}
	}

	return nil
}
