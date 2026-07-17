package redis

import (
	"context"
	"fmt"
	"sort"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

var enqueueOpsTaskOnceScript = goredis.NewScript(`
if redis.call("SET", KEYS[1], "1", "NX", "EX", ARGV[1]) == false then
  return ""
end
local fields = {}
for i = 2, #ARGV do
  fields[#fields + 1] = ARGV[i]
end
return redis.call("XADD", KEYS[2], "MAXLEN", "~", 1000, "*", unpack(fields))
`)

// EnqueueOpsTask 向诊断任务 Stream 追加一条任务
func EnqueueOpsTask(ctx context.Context, payload map[string]interface{}) (string, error) {
	id, err := RDB.XAdd(ctx, &goredis.XAddArgs{
		Stream: StreamOpsTasks,
		MaxLen: 1000,
		Approx: true,
		Values: payload,
	}).Result()
	if err != nil {
		return "", fmt.Errorf("stream: enqueue ops task: %w", err)
	}
	return id, nil
}

// EnqueueOpsTaskOnce 以 Redis Lua 原子完成去重标记和 XADD。
// 相同 dedupKey 在 ttl 内只会入队一次；返回 enqueued=false 表示已存在任务。
func EnqueueOpsTaskOnce(ctx context.Context, dedupKey string, ttl time.Duration, payload map[string]interface{}) (enqueued bool, messageID string, err error) {
	if dedupKey == "" {
		id, enqueueErr := EnqueueOpsTask(ctx, payload)
		return enqueueErr == nil, id, enqueueErr
	}
	if ttl <= 0 {
		ttl = 24 * time.Hour
	}

	keys := make([]string, 0, len(payload))
	for key := range payload {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	args := make([]interface{}, 0, 1+len(keys)*2)
	args = append(args, int64(ttl.Seconds()))
	for _, key := range keys {
		args = append(args, key, fmt.Sprint(payload[key]))
	}

	result, err := enqueueOpsTaskOnceScript.Run(ctx, RDB, []string{dedupKey, StreamOpsTasks}, args...).Text()
	if err != nil {
		return false, "", fmt.Errorf("stream: enqueue deduplicated ops task: %w", err)
	}
	if result == "" {
		return false, "", nil
	}
	return true, result, nil
}

// ReadOpsTasks 以 XREADGROUP 方式阻塞读取诊断任务
func ReadOpsTasks(ctx context.Context, consumer string, count int64, blockMs int64) ([]goredis.XMessage, error) {
	streams, err := RDB.XReadGroup(ctx, &goredis.XReadGroupArgs{
		Group:    GroupOpsWorker,
		Consumer: consumer,
		Streams:  []string{StreamOpsTasks, ">"},
		Count:    count,
		Block:    time.Duration(blockMs) * time.Millisecond,
		NoAck:    false,
	}).Result()
	if err != nil {
		if isNilOrTimeout(err) {
			return nil, nil
		}
		return nil, err
	}
	if len(streams) == 0 {
		return nil, nil
	}
	return streams[0].Messages, nil
}

// ClaimStaleOpsTasks 接管超过 minIdle 仍未 ACK 的诊断任务，用于 Worker 崩溃恢复。
func ClaimStaleOpsTasks(ctx context.Context, consumer string, count int64, minIdle time.Duration) ([]goredis.XMessage, error) {
	return claimStaleTasks(ctx, StreamOpsTasks, GroupOpsWorker, consumer, count, minIdle)
}

// AckOpsTask 确认诊断任务已处理
func AckOpsTask(ctx context.Context, msgID string) error {
	return RDB.XAck(ctx, StreamOpsTasks, GroupOpsWorker, msgID).Err()
}
