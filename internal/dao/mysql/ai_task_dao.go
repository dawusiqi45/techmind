package mysql

import (
	"techmind/internal/model"
)

// CreateAITask 创建任务记录
func CreateAITask(t *model.AITask) error {
	return DB.Create(t).Error
}

// UpdateAITaskStatus 更新任务状态
func UpdateAITaskStatus(id int64, status, errMsg string) error {
	return DB.Model(&model.AITask{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":    status,
			"error_msg": errMsg,
		}).Error
}

// IncrAITaskRetry 重试计数 +1，并设置为 pending 等待下次消费
func IncrAITaskRetry(id int64) error {
	return DB.Model(&model.AITask{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"retry_count": DB.Raw("retry_count + 1"),
			"status":      model.AITaskStatusPending,
		}).Error
}

// GetAITaskByTypeAndRef 按任务类型和关联 ID 查询最新任务
func GetAITaskByTypeAndRef(taskType string, refID int64) (*model.AITask, error) {
	var t model.AITask
	err := DB.Where("task_type = ? AND ref_id = ?", taskType, refID).
		Order("created_at DESC").First(&t).Error
	if err != nil {
		return nil, err
	}
	return &t, nil
}
