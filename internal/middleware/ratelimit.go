package middleware

import (
	"context"
	"fmt"
	"time"

	myredis "techmind/internal/dao/redis"
	"techmind/internal/pkg/response"
	"techmind/internal/pkg/settings"

	"github.com/gin-gonic/gin"
	goredis "github.com/redis/go-redis/v9"
)

func RateLimit(cfg *settings.RateLimitSetting) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !cfg.Enabled || myredis.RDB == nil {
			c.Next()
			return
		}

		ctx := context.Background()
		key := rateLimitKey(c)
		now := time.Now()
		windowStart := now.Add(-time.Minute).UnixMilli()

		pipe := myredis.RDB.Pipeline()
		pipe.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", windowStart))
		countCmd := pipe.ZCard(ctx, key)
		pipe.ZAdd(ctx, key, goredis.Z{
			Score:  float64(now.UnixMilli()),
			Member: fmt.Sprintf("%d", now.UnixNano()),
		})
		pipe.Expire(ctx, key, 2*time.Minute)
		_, err := pipe.Exec(ctx)

		if err != nil {
			c.Next()
			return
		}

		if countCmd.Val() >= int64(cfg.RequestsPerMin) {
			response.AbortWithRateLimited(c)
			return
		}

		c.Next()
	}
}

func rateLimitKey(c *gin.Context) string {
	if uid, ok := GetCurrentUserID(c); ok {
		return fmt.Sprintf("tm:ratelimit:user:%d", uid)
	}
	return fmt.Sprintf("tm:ratelimit:ip:%s", c.ClientIP())
}
