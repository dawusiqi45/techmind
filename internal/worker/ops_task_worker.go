package worker

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"techmind/internal/agent"
	redisDAO "techmind/internal/dao/redis"
	"techmind/internal/monitor"

	goredis "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// OpsWorker 消费诊断任务 Stream
type OpsWorker struct {
	consumer string
}

const (
	maxOpsRetry  = 3
	staleOpsIdle = 15 * time.Minute
)

// NewOpsWorker 创建 OpsWorker
func NewOpsWorker(consumer string) *OpsWorker {
	return &OpsWorker{consumer: consumer}
}

// Start 启动诊断任务消费循环
func (w *OpsWorker) Start(ctx context.Context) error {
	if err := redisDAO.EnsureConsumerGroup(ctx, redisDAO.StreamOpsTasks, redisDAO.GroupOpsWorker); err != nil {
		return fmt.Errorf("ops worker: ensure group: %w", err)
	}
	zap.L().Info("Ops worker started", zap.String("consumer", w.consumer))

	for {
		select {
		case <-ctx.Done():
			zap.L().Info("Ops worker stopping")
			return nil
		default:
		}

		msgs, err := redisDAO.ClaimStaleOpsTasks(ctx, w.consumer, 5, staleOpsIdle)
		if err != nil {
			zap.L().Warn("ops worker: claim stale tasks failed", zap.Error(err))
		}
		if len(msgs) == 0 {
			msgs, err = redisDAO.ReadOpsTasks(ctx, w.consumer, 5, 2000)
		}
		if err != nil {
			zap.L().Warn("ops worker: read error", zap.Error(err))
			time.Sleep(1 * time.Second)
			continue
		}

		for _, msg := range msgs {
			startedAt := time.Now()
			w.processOps(ctx, msg)
			monitor.ObserveRedisStreamConsume(redisDAO.StreamOpsTasks, redisDAO.GroupOpsWorker, time.Since(startedAt))
		}
		if pending, err := redisDAO.PendingOpsTasks(ctx); err == nil {
			monitor.SetRedisStreamPending(redisDAO.StreamOpsTasks, redisDAO.GroupOpsWorker, pending)
		}
		if length, err := redisDAO.StreamLen(ctx, redisDAO.StreamOpsTasks); err == nil {
			monitor.SetRedisStreamLen(redisDAO.StreamOpsTasks, length)
		}
	}
}

func (w *OpsWorker) processOps(ctx context.Context, msg goredis.XMessage) {
	alertIDStr, _ := msg.Values["alert_id"].(string)
	triggerType, _ := msg.Values["trigger_type"].(string)
	service, _ := msg.Values["service"].(string)
	alertName, _ := msg.Values["alert_name"].(string)

	alertID, _ := strconv.ParseInt(alertIDStr, 10, 64)

	input := agent.DiagnoseInput{
		AlertID:     alertID,
		TriggerType: triggerType,
		Service:     service,
		AlertName:   alertName,
	}

	if _, err := agent.Diagnose(ctx, input); err != nil {
		zap.L().Warn("ops worker: diagnose failed",
			zap.String("msg_id", msg.ID),
			zap.Error(err))
		monitor.IncWorkerTask("ops_diagnose", "failed")

		retryCount := parseRetry(msg.Values["retry_count"])
		if retryCount >= maxOpsRetry {
			deadPayload := copyValues(msg.Values)
			deadPayload["fail_reason"] = err.Error()
			deadPayload["original_msg_id"] = msg.ID
			if deadErr := redisDAO.EnqueueOpsDeadLetter(ctx, deadPayload); deadErr != nil {
				zap.L().Error("ops worker: enqueue dead letter failed; leaving original pending", zap.Error(deadErr))
				return
			}
			if ackErr := redisDAO.AckOpsTask(ctx, msg.ID); ackErr != nil {
				zap.L().Error("ops worker: ack dead-lettered task failed", zap.Error(ackErr))
				return
			}
			monitor.IncWorkerTask("ops_diagnose", "dead")
		} else {
			retryPayload := copyValues(msg.Values)
			retryPayload["retry_count"] = strconv.Itoa(retryCount + 1)
			if _, enqueueErr := redisDAO.EnqueueOpsTask(ctx, retryPayload); enqueueErr != nil {
				zap.L().Error("ops worker: re-enqueue failed; leaving original pending", zap.Error(enqueueErr))
				return
			} else {
				if ackErr := redisDAO.AckOpsTask(ctx, msg.ID); ackErr != nil {
					zap.L().Error("ops worker: ack re-enqueued task failed", zap.Error(ackErr))
					return
				}
				monitor.IncWorkerTask("ops_diagnose", "retry")
			}
		}
	} else {
		if ackErr := redisDAO.AckOpsTask(ctx, msg.ID); ackErr != nil {
			zap.L().Error("ops worker: ack completed task failed", zap.Error(ackErr))
			return
		}
		monitor.IncWorkerTask("ops_diagnose", "success")
	}
}

// EnqueueDiagnoseTask 将诊断任务入队 ops_tasks Stream
func EnqueueDiagnoseTask(ctx context.Context, alertID int64, triggerType, service, alertName string) error {
	payload := map[string]interface{}{
		"alert_id":     strconv.FormatInt(alertID, 10),
		"trigger_type": triggerType,
		"service":      service,
		"alert_name":   alertName,
		"retry_count":  "0",
	}
	_, err := redisDAO.EnqueueOpsTask(ctx, payload)
	return err
}
