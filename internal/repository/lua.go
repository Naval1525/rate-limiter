package repository

import (
	"context"
	_ "embed"
	"time"

	"github.com/redis/go-redis/v9"
)

//go:embed lua/token_bucket.lua
var tokenBucketScript string

type LuaScript struct {
	Script *redis.Script
}

func NewLuaScript() *LuaScript {
	return &LuaScript{
		Script: redis.NewScript(tokenBucketScript),
	}
}

func (l *LuaScript) Allow(ctx context.Context, rdb *redis.Client, key string, capacity, rate float64) (bool, error) {
	now := time.Now().UnixMilli()

	result, err := l.Script.Run(ctx, rdb, []string{key},
		capacity,
		rate,
		now,
	).Int()

	if err != nil {
		return false, err
	}

	return result == 1, nil
}
