package worker

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"techmind/internal/agent"
	redisDAO "techmind/internal/dao/redis"

	goredis "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// OpsWorker 消费诊断任务 Stream
type OpsWorker struct {
	consumer string
}

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

		msgs, err := redisDAO.ReadOpsTasks(ctx, w.consumer, 5, 2000)
		if err != nil {
			zap.L().Warn("ops worker: read error", zap.Error(err))
			time.Sleep(1 * time.Second)
			continue
		}

		for _, msg := range msgs {
			w.processOps(ctx, msg)
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
	}
	_ = redisDAO.AckOpsTask(ctx, msg.ID)
}

// EnqueueDiagnoseTask 将诊断任务入队 ops_tasks Stream
func EnqueueDiagnoseTask(ctx context.Context, alertID int64, triggerType, service, alertName string) error {
	payload := map[string]interface{}{
		"alert_id":     strconv.FormatInt(alertID, 10),
		"trigger_type": triggerType,
		"service":      service,
		"alert_name":   alertName,
	}
	_, err := redisDAO.EnqueueOpsTask(ctx, payload)
	return err
}
