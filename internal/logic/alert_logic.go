package logic

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"strings"
	"time"

	"techmind/internal/alert"
	mysqlDAO "techmind/internal/dao/mysql"
	"techmind/internal/model"
	"techmind/internal/pkg/snowflake"
)

// AlertmanagerPayload 是 Alertmanager webhook POST 的顶层结构
type AlertmanagerPayload struct {
	Receiver string              `json:"receiver"`
	Status   string              `json:"status"` // firing / resolved
	Alerts   []AlertmanagerAlert `json:"alerts"`
}

// AlertmanagerAlert 是单条告警
type AlertmanagerAlert struct {
	Status      string            `json:"status"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	StartsAt    time.Time         `json:"startsAt"`
	EndsAt      time.Time         `json:"endsAt"`
}

// AlertDetail 告警详情（含增强上下文）
type AlertDetail struct {
	*model.AlertEvent
	Enrichment *model.AlertEnrichment `json:"enrichment,omitempty"`
}

// ReceiveAlertWebhook 解析 Alertmanager webhook，对每条告警去重写库
func ReceiveAlertWebhook(payload *AlertmanagerPayload) error {
	var errs []error
	for _, a := range payload.Alerts {
		if err := upsertAlert(a); err != nil {
			errs = append(errs, fmt.Errorf("upsert alert %q: %w", a.Labels["alertname"], err))
		}
	}
	return errors.Join(errs...)
}

func upsertAlert(a AlertmanagerAlert) error {
	alertName := a.Labels["alertname"]
	service := a.Labels["service"]
	endpoint := a.Labels["endpoint"]
	severity := a.Labels["severity"]
	if severity == "" {
		severity = "warning"
	}

	fingerprint := buildFingerprint(alertName, service, endpoint, severity)

	status := model.AlertStatusFiring
	if a.Status == "resolved" {
		status = model.AlertStatusResolved
	}

	labels := make(model.JSONMap, len(a.Labels))
	for k, v := range a.Labels {
		labels[k] = v
	}
	annotations := make(model.JSONMap, len(a.Annotations))
	for k, v := range a.Annotations {
		annotations[k] = v
	}

	event := &model.AlertEvent{
		ID:          snowflake.GenID(),
		Fingerprint: fingerprint,
		AlertName:   alertName,
		Service:     service,
		Endpoint:    endpoint,
		Severity:    severity,
		Status:      status,
		Labels:      labels,
		Annotations: annotations,
		FirstSeenAt: a.StartsAt,
		LastSeenAt:  time.Now(),
	}
	if err := mysqlDAO.UpsertAlertEvent(event); err != nil {
		return err
	}

	// 异步执行告警增强，失败不阻断
	go alert.EnrichAlert(context.Background(), event)
	return nil
}

// buildFingerprint 按 alert_name + service + endpoint + severity 生成唯一指纹
func buildFingerprint(alertName, service, endpoint, severity string) string {
	raw := strings.Join([]string{alertName, service, endpoint, severity}, "|")
	hash := sha256.Sum256([]byte(raw))
	return fmt.Sprintf("%x", hash[:8]) // 16 chars
}

// ListAlerts 分页查询告警列表
func ListAlerts(status string, page, pageSize int) ([]*model.AlertEvent, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return mysqlDAO.ListAlertEvents(status, page, pageSize)
}

// GetAlertDetail 获取告警详情（含增强上下文）
func GetAlertDetail(id int64) (*AlertDetail, error) {
	event, err := mysqlDAO.GetAlertEventByID(id)
	if err != nil {
		return nil, err
	}
	if event == nil {
		return nil, nil
	}
	enrichment, _ := mysqlDAO.GetAlertEnrichmentByAlertID(id)
	return &AlertDetail{AlertEvent: event, Enrichment: enrichment}, nil
}

// AcknowledgeAlert 确认告警
func AcknowledgeAlert(id int64) error {
	event, err := mysqlDAO.GetAlertEventByID(id)
	if err != nil || event == nil {
		return fmt.Errorf("alert not found")
	}
	return mysqlDAO.AcknowledgeAlert(id)
}
