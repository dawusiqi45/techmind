package redis

import (
	"context"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

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
