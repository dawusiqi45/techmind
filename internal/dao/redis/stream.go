package redis

import (
	"context"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

const (
	StreamAITasks    = "tm:stream:ai_tasks"      // AI 异步任务队列
	StreamAIDeadLetter = "tm:stream:ai_tasks:dead" // AI 死信队列
	StreamOpsTasks   = "tm:stream:ops_tasks"     // 诊断任务队列

	GroupAIWorker  = "ai_worker_group"
	GroupOpsWorker = "ops_worker_group"

	// 任务类型常量，供 Logic 和 Worker 共用
	TaskArticleSummary     = "article.summary"
	TaskArticleTag         = "article.tag"
	TaskArticleIndex       = "article.index"
	TaskArticleReindex     = "article.reindex"
	TaskArticleDeleteIndex = "article.delete_index"
)

// EnqueueAITask 向 AI 任务 Stream 追加一条任务消息
// payload 为 map[string]interface{}，例如：{"task_type": "article.summary", "article_id": "123"}
func EnqueueAITask(ctx context.Context, payload map[string]interface{}) (string, error) {
	id, err := RDB.XAdd(ctx, &goredis.XAddArgs{
		Stream: StreamAITasks,
		MaxLen: 10000,
		Approx: true,
		Values: payload,
	}).Result()
	if err != nil {
		return "", fmt.Errorf("stream: enqueue ai task: %w", err)
	}
	return id, nil
}

// EnqueueDeadLetter 将失败消息转入死信队列
func EnqueueDeadLetter(ctx context.Context, payload map[string]interface{}) error {
	_, err := RDB.XAdd(ctx, &goredis.XAddArgs{
		Stream: StreamAIDeadLetter,
		MaxLen: 5000,
		Approx: true,
		Values: payload,
	}).Result()
	return err
}

// EnsureConsumerGroup 创建消费者组（已存在则忽略）
func EnsureConsumerGroup(ctx context.Context, stream, group string) error {
	err := RDB.XGroupCreateMkStream(ctx, stream, group, "0").Err()
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		return fmt.Errorf("stream: create group %q on %q: %w", group, stream, err)
	}
	return nil
}

// ReadAITasks 以 XREADGROUP 方式阻塞读取 AI 任务，count 条，阻塞 blockMs 毫秒
func ReadAITasks(ctx context.Context, consumer string, count int64, blockMs int64) ([]goredis.XMessage, error) {
	streams, err := RDB.XReadGroup(ctx, &goredis.XReadGroupArgs{
		Group:    GroupAIWorker,
		Consumer: consumer,
		Streams:  []string{StreamAITasks, ">"},
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

// AckAITask 确认消息已成功处理
func AckAITask(ctx context.Context, msgID string) error {
	return RDB.XAck(ctx, StreamAITasks, GroupAIWorker, msgID).Err()
}

// PendingAITasks 查询当前 pending（已读未 ACK）消息数量
func PendingAITasks(ctx context.Context) (int64, error) {
	info, err := RDB.XPending(ctx, StreamAITasks, GroupAIWorker).Result()
	if err != nil {
		return 0, err
	}
	return info.Count, nil
}

// StreamLen 返回 Stream 当前长度（用于观测）
func StreamLen(ctx context.Context, stream string) (int64, error) {
	return RDB.XLen(ctx, stream).Result()
}

// isNilOrTimeout 判断是否为空结果或超时（XREADGROUP block 正常超时）
func isNilOrTimeout(err error) bool {
	if err == nil {
		return true
	}
	s := err.Error()
	return s == "redis: nil" || s == "EOF"
}
