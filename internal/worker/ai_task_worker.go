package worker

import (
	"context"
	"fmt"
	"strconv"
	"time"

	redisDAO "techmind/internal/dao/redis"
	"techmind/internal/model"
	"techmind/internal/monitor"
	"techmind/internal/pkg/snowflake"

	mysqlDAO "techmind/internal/dao/mysql"

	goredis "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const (
	maxRetry      = 3    // 最大重试次数，超过后转死信
	pollBatchSize = 10   // 每次 XREADGROUP 拉取条数
	blockMs       = 2000 // 阻塞等待毫秒数
	staleAIIdle   = 2 * time.Minute
)

// TaskHandler 任务处理函数签名
type TaskHandler func(ctx context.Context, msg goredis.XMessage) error

// AIWorker 消费 AI 任务 Stream
type AIWorker struct {
	consumer string                 // 消费者名称（唯一）
	handlers map[string]TaskHandler // task_type → handler
}

// NewAIWorker 创建 AIWorker，consumer 建议用主机名或 pod 名
func NewAIWorker(consumer string) *AIWorker {
	return &AIWorker{
		consumer: consumer,
		handlers: make(map[string]TaskHandler),
	}
}

// Register 注册任务处理函数
func (w *AIWorker) Register(taskType string, h TaskHandler) {
	w.handlers[taskType] = h
}

// Start 启动消费循环，ctx 取消时退出
func (w *AIWorker) Start(ctx context.Context) error {
	// 确保消费者组存在
	if err := redisDAO.EnsureConsumerGroup(ctx, redisDAO.StreamAITasks, redisDAO.GroupAIWorker); err != nil {
		return fmt.Errorf("ai worker: ensure group: %w", err)
	}
	zap.L().Info("AI worker started", zap.String("consumer", w.consumer))

	for {
		select {
		case <-ctx.Done():
			zap.L().Info("AI worker stopping")
			return nil
		default:
		}

		msgs, err := redisDAO.ClaimStaleAITasks(ctx, w.consumer, pollBatchSize, staleAIIdle)
		if err != nil {
			zap.L().Warn("AI worker: claim stale tasks failed", zap.Error(err))
		}
		if len(msgs) == 0 {
			msgs, err = redisDAO.ReadAITasks(ctx, w.consumer, pollBatchSize, blockMs)
		}
		if err != nil {
			zap.L().Warn("AI worker: read stream error", zap.Error(err))
			time.Sleep(1 * time.Second)
			continue
		}

		for _, msg := range msgs {
			startedAt := time.Now()
			w.process(ctx, msg)
			monitor.ObserveRedisStreamConsume(redisDAO.StreamAITasks, redisDAO.GroupAIWorker, time.Since(startedAt))
		}
		if pending, err := redisDAO.PendingAITasks(ctx); err == nil {
			monitor.SetRedisStreamPending(redisDAO.StreamAITasks, redisDAO.GroupAIWorker, pending)
		}
		if length, err := redisDAO.StreamLen(ctx, redisDAO.StreamAITasks); err == nil {
			monitor.SetRedisStreamLen(redisDAO.StreamAITasks, length)
		}
	}
}

// process 处理单条消息：成功→ACK，失败→重试或死信
func (w *AIWorker) process(ctx context.Context, msg goredis.XMessage) {
	taskType, _ := msg.Values["task_type"].(string)
	taskIDStr, _ := msg.Values["task_id"].(string)

	log := zap.L().With(
		zap.String("msg_id", msg.ID),
		zap.String("task_type", taskType),
		zap.String("task_id", taskIDStr),
	)

	handler, ok := w.handlers[taskType]
	if !ok {
		deadPayload := copyValues(msg.Values)
		deadPayload["fail_reason"] = "no handler for task type"
		deadPayload["original_msg_id"] = msg.ID
		if err := redisDAO.EnqueueDeadLetter(ctx, deadPayload); err != nil {
			log.Error("AI worker: move unknown task to dead letter failed", zap.Error(err))
			return
		}
		if err := redisDAO.AckAITask(ctx, msg.ID); err != nil {
			log.Error("AI worker: ack unknown task failed", zap.Error(err))
		}
		return
	}

	taskID := parseID(taskIDStr)
	if taskID > 0 {
		_ = mysqlDAO.UpdateAITaskStatus(taskID, model.AITaskStatusRunning, "")
	}

	err := handler(ctx, msg)
	if err == nil {
		if ackErr := redisDAO.AckAITask(ctx, msg.ID); ackErr != nil {
			log.Error("AI worker: ack completed task failed", zap.Error(ackErr))
			return
		}
		monitor.IncWorkerTask(taskType, "success")
		if taskID > 0 {
			_ = mysqlDAO.UpdateAITaskStatus(taskID, model.AITaskStatusDone, "")
		}
		log.Info("AI worker: task done")
		return
	}

	// 处理失败
	log.Warn("AI worker: task failed", zap.Error(err))
	monitor.IncWorkerTask(taskType, "failed")

	retryCount := parseRetry(msg.Values["retry_count"])
	if retryCount >= maxRetry {
		// 超过重试次数，转死信
		deadPayload := copyValues(msg.Values)
		deadPayload["fail_reason"] = err.Error()
		deadPayload["original_msg_id"] = msg.ID
		if deadErr := redisDAO.EnqueueDeadLetter(ctx, deadPayload); deadErr != nil {
			log.Error("AI worker: enqueue dead letter failed; leaving original pending", zap.Error(deadErr))
			return
		}
		if ackErr := redisDAO.AckAITask(ctx, msg.ID); ackErr != nil {
			log.Error("AI worker: ack dead-lettered task failed", zap.Error(ackErr))
			return
		}
		if taskID > 0 {
			_ = mysqlDAO.UpdateAITaskStatus(taskID, model.AITaskStatusDead, err.Error())
		}
		monitor.IncWorkerTask(taskType, "dead")
		log.Error("AI worker: task moved to dead letter", zap.Int("retry_count", retryCount))
		return
	}

	// 重新入队（retry_count +1）
	retryPayload := copyValues(msg.Values)
	retryPayload["retry_count"] = strconv.Itoa(retryCount + 1)
	if _, enqErr := redisDAO.EnqueueAITask(ctx, retryPayload); enqErr != nil {
		log.Error("AI worker: re-enqueue failed; leaving original pending", zap.Error(enqErr))
		return
	}
	if ackErr := redisDAO.AckAITask(ctx, msg.ID); ackErr != nil {
		log.Error("AI worker: ack re-enqueued task failed", zap.Error(ackErr))
		return
	}
	if taskID > 0 {
		_ = mysqlDAO.IncrAITaskRetry(taskID)
	}
	monitor.IncWorkerTask(taskType, "retry")
	log.Warn("AI worker: task re-enqueued for retry", zap.Int("retry_count", retryCount+1))
}

// EnqueueTask 入队辅助函数，供 logic 层调用
// taskType 如 "article.summary"，refID 如 articleID
func EnqueueTask(ctx context.Context, taskType string, refID int64, extra map[string]interface{}) error {
	taskID := snowflake.GenID()

	// 写 DB 记录
	t := &model.AITask{
		ID:       taskID,
		TaskType: taskType,
		RefID:    refID,
		Status:   model.AITaskStatusPending,
	}
	if err := mysqlDAO.CreateAITask(t); err != nil {
		return fmt.Errorf("enqueue: create ai_task: %w", err)
	}

	// 入 Stream
	payload := map[string]interface{}{
		"task_type":   taskType,
		"task_id":     strconv.FormatInt(taskID, 10),
		"ref_id":      strconv.FormatInt(refID, 10),
		"retry_count": "0",
	}
	for k, v := range extra {
		payload[k] = v
	}
	if _, err := redisDAO.EnqueueAITask(ctx, payload); err != nil {
		_ = mysqlDAO.UpdateAITaskStatus(taskID, model.AITaskStatusFailed, err.Error())
		return fmt.Errorf("enqueue: xadd: %w", err)
	}
	return nil
}

// parseID 安全解析 int64 ID
func parseID(s string) int64 {
	id, _ := strconv.ParseInt(s, 10, 64)
	return id
}

// parseRetry 安全解析重试次数
func parseRetry(v interface{}) int {
	switch val := v.(type) {
	case string:
		n, _ := strconv.Atoi(val)
		return n
	case int:
		return val
	}
	return 0
}

// copyValues 浅拷贝消息 Values map
func copyValues(src map[string]interface{}) map[string]interface{} {
	dst := make(map[string]interface{}, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}
