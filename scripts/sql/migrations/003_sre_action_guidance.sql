-- SRE Agent 结构化操作建议升级。
-- 可重复执行：增加只读排查、修改方案、验证和回滚四类 JSON 字段。

SET @schema_name = DATABASE();

SET @has_verification_commands = (
    SELECT COUNT(*) FROM information_schema.columns
    WHERE table_schema = @schema_name AND table_name = 'ops_report' AND column_name = 'verification_commands'
);
SET @verification_commands_ddl = IF(
    @has_verification_commands = 0,
    'ALTER TABLE `ops_report` ADD COLUMN `verification_commands` JSON NULL AFTER `suggestions`',
    'SELECT 1'
);
PREPARE verification_commands_stmt FROM @verification_commands_ddl;
EXECUTE verification_commands_stmt;
DEALLOCATE PREPARE verification_commands_stmt;
UPDATE `ops_report` SET `verification_commands` = JSON_ARRAY() WHERE `verification_commands` IS NULL;
ALTER TABLE `ops_report` MODIFY COLUMN `verification_commands` JSON NOT NULL COMMENT '只读排查命令';

SET @has_change_plan = (
    SELECT COUNT(*) FROM information_schema.columns
    WHERE table_schema = @schema_name AND table_name = 'ops_report' AND column_name = 'change_plan'
);
SET @change_plan_ddl = IF(
    @has_change_plan = 0,
    'ALTER TABLE `ops_report` ADD COLUMN `change_plan` JSON NULL AFTER `verification_commands`',
    'SELECT 1'
);
PREPARE change_plan_stmt FROM @change_plan_ddl;
EXECUTE change_plan_stmt;
DEALLOCATE PREPARE change_plan_stmt;
UPDATE `ops_report` SET `change_plan` = JSON_ARRAY() WHERE `change_plan` IS NULL;
ALTER TABLE `ops_report` MODIFY COLUMN `change_plan` JSON NOT NULL COMMENT '需人工审批的修改方案';

SET @has_validation_commands = (
    SELECT COUNT(*) FROM information_schema.columns
    WHERE table_schema = @schema_name AND table_name = 'ops_report' AND column_name = 'validation_commands'
);
SET @validation_commands_ddl = IF(
    @has_validation_commands = 0,
    'ALTER TABLE `ops_report` ADD COLUMN `validation_commands` JSON NULL AFTER `change_plan`',
    'SELECT 1'
);
PREPARE validation_commands_stmt FROM @validation_commands_ddl;
EXECUTE validation_commands_stmt;
DEALLOCATE PREPARE validation_commands_stmt;
UPDATE `ops_report` SET `validation_commands` = JSON_ARRAY() WHERE `validation_commands` IS NULL;
ALTER TABLE `ops_report` MODIFY COLUMN `validation_commands` JSON NOT NULL COMMENT '修改后验证命令';

SET @has_rollback_commands = (
    SELECT COUNT(*) FROM information_schema.columns
    WHERE table_schema = @schema_name AND table_name = 'ops_report' AND column_name = 'rollback_commands'
);
SET @rollback_commands_ddl = IF(
    @has_rollback_commands = 0,
    'ALTER TABLE `ops_report` ADD COLUMN `rollback_commands` JSON NULL AFTER `validation_commands`',
    'SELECT 1'
);
PREPARE rollback_commands_stmt FROM @rollback_commands_ddl;
EXECUTE rollback_commands_stmt;
DEALLOCATE PREPARE rollback_commands_stmt;
UPDATE `ops_report` SET `rollback_commands` = JSON_ARRAY() WHERE `rollback_commands` IS NULL;
ALTER TABLE `ops_report` MODIFY COLUMN `rollback_commands` JSON NOT NULL COMMENT '回滚命令';
