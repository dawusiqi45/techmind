-- SRE Agent 自动诊断与 Worker 重试幂等升级。
-- 可重复执行：为 ops_report 增加任务幂等键，并兼容已有报告。

SET @schema_name = DATABASE();
SET @has_task_key = (
    SELECT COUNT(*) FROM information_schema.columns
    WHERE table_schema = @schema_name
      AND table_name = 'ops_report'
      AND column_name = 'task_key'
);
SET @task_key_ddl = IF(
    @has_task_key = 0,
    'ALTER TABLE `ops_report` ADD COLUMN `task_key` VARCHAR(128) NULL AFTER `trigger_type`',
    'SELECT 1'
);
PREPARE task_key_stmt FROM @task_key_ddl;
EXECUTE task_key_stmt;
DEALLOCATE PREPARE task_key_stmt;

UPDATE `ops_report`
SET `task_key` = CONCAT('legacy:', `id`)
WHERE `task_key` IS NULL OR `task_key` = '';

SET @has_task_key_index = (
    SELECT COUNT(*) FROM information_schema.statistics
    WHERE table_schema = @schema_name
      AND table_name = 'ops_report'
      AND index_name = 'uk_task_key'
);
SET @task_key_index_ddl = IF(
    @has_task_key_index = 0,
    'ALTER TABLE `ops_report` MODIFY COLUMN `task_key` VARCHAR(128) NOT NULL, ADD UNIQUE KEY `uk_task_key` (`task_key`)',
    'SELECT 1'
);
PREPARE task_key_index_stmt FROM @task_key_index_ddl;
EXECUTE task_key_index_stmt;
DEALLOCATE PREPARE task_key_index_stmt;
