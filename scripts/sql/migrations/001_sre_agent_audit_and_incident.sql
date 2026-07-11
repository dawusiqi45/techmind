-- TechMind 现有 MySQL PVC 的兼容升级脚本。
-- 可重复执行：为诊断证据链补表，并把诊断报告关联到 Incident。

CREATE TABLE IF NOT EXISTS `ops_tool_call` (
    `id`          BIGINT       NOT NULL AUTO_INCREMENT,
    `report_id`   BIGINT       NOT NULL,
    `tool_name`   VARCHAR(64)  NOT NULL,
    `input`       JSON         NOT NULL,
    `output`      JSON         NOT NULL,
    `duration_ms` INT          NOT NULL DEFAULT 0,
    `created_at`  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_report_id` (`report_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Agent工具调用审计表';

SET @schema_name = DATABASE();
SET @has_incident_id = (
    SELECT COUNT(*)
    FROM information_schema.columns
    WHERE table_schema = @schema_name
      AND table_name = 'ops_report'
      AND column_name = 'incident_id'
);
SET @incident_ddl = IF(
    @has_incident_id = 0,
    'ALTER TABLE `ops_report` ADD COLUMN `incident_id` BIGINT NOT NULL DEFAULT 0 COMMENT ''关联故障事件ID，手动诊断为0'' AFTER `alert_id`, ADD KEY `idx_incident_id` (`incident_id`)',
    'SELECT 1'
);
PREPARE incident_stmt FROM @incident_ddl;
EXECUTE incident_stmt;
DEALLOCATE PREPARE incident_stmt;
