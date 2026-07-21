package middleware

import (
	"fmt"
	"strconv"
	"time"

	myredis "techmind/internal/dao/redis"
	"techmind/internal/pkg/response"
	"techmind/internal/pkg/settings"

	"github.com/gin-gonic/gin"
)

var rateLimitScript = `
redis.call('ZREMRANGEBYSCORE', KEYS[1], 0, ARGV[1])
local count = redis.call('ZCARD', KEYS[1])
if count >= tonumber(ARGV[3]) then
  redis.call('EXPIRE', KEYS[1], 120)
  return 0
end
redis.call('ZADD', KEYS[1], ARGV[2], ARGV[4])
redis.call('EXPIRE', KEYS[1], 120)
return 1
`

func RateLimit(cfg *settings.RateLimitSetting) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !cfg.Enabled || myredis.RDB == nil {
			c.Next()
			return
		}

		ctx := c.Request.Context()
		key := rateLimitKey(c)
		now := time.Now()
		windowStart := now.Add(-time.Minute).UnixMilli()
		allowed, err := myredis.RDB.Eval(ctx, rateLimitScript, []string{key},
			strconv.FormatInt(windowStart, 10), strconv.FormatInt(now.UnixMilli(), 10),
			strconv.Itoa(cfg.RequestsPerMin), strconv.FormatInt(now.UnixNano(), 10)).Int()
		if err != nil {
			c.Next()
			return
		}
		if allowed != 1 {
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
