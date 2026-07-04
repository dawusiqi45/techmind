-- TechMind 数据库初始化脚本
-- 执行前确保数据库已创建：CREATE DATABASE techmind CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ============================================================
-- 业务表
-- ============================================================

CREATE TABLE IF NOT EXISTS `user` (
    `id`           BIGINT       NOT NULL COMMENT '雪花ID',
    `username`     VARCHAR(64)  NOT NULL COMMENT '用户名',
    `password`     VARCHAR(128) NOT NULL COMMENT 'bcrypt 密码哈希',
    `email`        VARCHAR(128) NOT NULL DEFAULT '' COMMENT '邮箱',
    `avatar`       VARCHAR(256) NOT NULL DEFAULT '' COMMENT '头像URL',
    `role`         TINYINT      NOT NULL DEFAULT 0 COMMENT '0=普通用户 1=管理员',
    `status`       TINYINT      NOT NULL DEFAULT 1 COMMENT '1=正常 0=禁用',
    `created_at`   DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`   DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_username` (`username`),
    UNIQUE KEY `uk_email` (`email`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';

CREATE TABLE IF NOT EXISTS `article` (
    `id`           BIGINT        NOT NULL COMMENT '雪花ID',
    `author_id`    BIGINT        NOT NULL COMMENT '作者ID',
    `title`        VARCHAR(256)  NOT NULL COMMENT '标题',
    `content`      MEDIUMTEXT    NOT NULL COMMENT 'Markdown 正文',
    `summary`      VARCHAR(512)  NOT NULL DEFAULT '' COMMENT 'AI 生成摘要',
    `cover`        VARCHAR(256)  NOT NULL DEFAULT '' COMMENT '封面图URL',
    `status`       TINYINT       NOT NULL DEFAULT 1 COMMENT '1=发布 0=草稿 -1=已删除',
    `index_status` TINYINT       NOT NULL DEFAULT 0 COMMENT '0=未索引 1=已索引 -1=索引失败',
    `view_count`   INT           NOT NULL DEFAULT 0,
    `like_count`   INT           NOT NULL DEFAULT 0,
    `favorite_count` INT         NOT NULL DEFAULT 0,
    `comment_count`  INT         NOT NULL DEFAULT 0,
    `created_at`   DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`   DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_author_id` (`author_id`),
    KEY `idx_created_at` (`created_at`),
    KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='文章表';

CREATE TABLE IF NOT EXISTS `tag` (
    `id`         BIGINT      NOT NULL COMMENT '雪花ID',
    `name`       VARCHAR(64) NOT NULL COMMENT '标签名',
    `hot_score`  FLOAT       NOT NULL DEFAULT 0 COMMENT '热度分',
    `created_at` DATETIME    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='标签表';

CREATE TABLE IF NOT EXISTS `article_tag` (
    `id`         BIGINT      NOT NULL AUTO_INCREMENT,
    `article_id` BIGINT      NOT NULL,
    `tag_id`     BIGINT      NOT NULL,
    `source`     VARCHAR(16) NOT NULL DEFAULT 'manual' COMMENT 'manual/ai',
    `created_at` DATETIME    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_article_tag` (`article_id`, `tag_id`),
    KEY `idx_tag_id` (`tag_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='文章标签关联表';

CREATE TABLE IF NOT EXISTS `comment` (
    `id`         BIGINT     NOT NULL COMMENT '雪花ID',
    `article_id` BIGINT     NOT NULL,
    `author_id`  BIGINT     NOT NULL,
    `parent_id`  BIGINT     NOT NULL DEFAULT 0 COMMENT '0=一级评论，否则为父评论ID',
    `content`    TEXT       NOT NULL,
    `status`     TINYINT    NOT NULL DEFAULT 1 COMMENT '1=正常 -1=已删除',
    `created_at` DATETIME   NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME   NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_article_id` (`article_id`),
    KEY `idx_parent_id` (`parent_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='评论表';

CREATE TABLE IF NOT EXISTS `favorite` (
    `id`         BIGINT   NOT NULL AUTO_INCREMENT,
    `user_id`    BIGINT   NOT NULL,
    `article_id` BIGINT   NOT NULL,
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_user_article` (`user_id`, `article_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='收藏表';

CREATE TABLE IF NOT EXISTS `article_chunk` (
    `id`         BIGINT       NOT NULL AUTO_INCREMENT,
    `article_id` BIGINT       NOT NULL,
    `chunk_index` INT         NOT NULL DEFAULT 0 COMMENT 'chunk 序号',
    `content`    TEXT         NOT NULL COMMENT 'chunk 正文',
    `milvus_id`  VARCHAR(64)  NOT NULL DEFAULT '' COMMENT 'Milvus 向量ID',
    `created_at` DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_article_id` (`article_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='文章向量chunk表';

-- ============================================================
-- AI 与任务表
-- ============================================================

CREATE TABLE IF NOT EXISTS `ai_task` (
    `id`          BIGINT       NOT NULL COMMENT '雪花ID',
    `task_type`   VARCHAR(32)  NOT NULL COMMENT 'article.summary / article.tag / article.index 等',
    `ref_id`      BIGINT       NOT NULL COMMENT '关联业务ID（如 article_id）',
    `status`      VARCHAR(16)  NOT NULL DEFAULT 'pending' COMMENT 'pending/running/done/failed/dead',
    `retry_count` INT          NOT NULL DEFAULT 0,
    `error_msg`   VARCHAR(512) NOT NULL DEFAULT '',
    `created_at`  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_task_type_status` (`task_type`, `status`),
    KEY `idx_ref_id` (`ref_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='AI异步任务表';

CREATE TABLE IF NOT EXISTS `ai_call_record` (
    `id`           BIGINT       NOT NULL COMMENT '雪花ID',
    `skill`        VARCHAR(64)  NOT NULL COMMENT 'Skill 名称',
    `model`        VARCHAR(64)  NOT NULL DEFAULT '',
    `input_tokens` INT          NOT NULL DEFAULT 0,
    `output_tokens` INT         NOT NULL DEFAULT 0,
    `duration_ms`  INT          NOT NULL DEFAULT 0,
    `status`       VARCHAR(16)  NOT NULL DEFAULT 'ok' COMMENT 'ok/failed',
    `error_msg`    VARCHAR(256) NOT NULL DEFAULT '',
    `ref_id`       BIGINT       NOT NULL DEFAULT 0 COMMENT '关联业务ID',
    `created_at`   DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_skill` (`skill`),
    KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='AI调用记录表';

CREATE TABLE IF NOT EXISTS `runbook` (
    `id`          BIGINT       NOT NULL COMMENT '雪花ID',
    `title`       VARCHAR(256) NOT NULL,
    `content`     MEDIUMTEXT   NOT NULL COMMENT 'Markdown 内容',
    `alert_name`  VARCHAR(128) NOT NULL DEFAULT '' COMMENT '关联告警名称',
    `service`     VARCHAR(64)  NOT NULL DEFAULT '',
    `created_at`  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='运维手册表';

CREATE TABLE IF NOT EXISTS `ops_report` (
    `id`           BIGINT       NOT NULL COMMENT '雪花ID',
    `alert_id`     BIGINT       NOT NULL DEFAULT 0 COMMENT '触发来源告警ID，手动触发为0',
    `trigger_type` VARCHAR(16)  NOT NULL DEFAULT 'manual' COMMENT 'manual/alert',
    `summary`      TEXT         NOT NULL,
    `evidence`     JSON         NOT NULL COMMENT '证据列表',
    `root_cause`   TEXT         NOT NULL,
    `impact`       TEXT         NOT NULL DEFAULT '',
    `suggestions`  JSON         NOT NULL COMMENT '建议列表',
    `related_changes` JSON      NOT NULL COMMENT '关联变更',
    `tool_calls`   JSON         NOT NULL COMMENT '工具调用记录',
    `status`       VARCHAR(16)  NOT NULL DEFAULT 'done',
    `created_at`   DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_alert_id` (`alert_id`),
    KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='SRE诊断报告表';

-- ============================================================
-- 可观测与告警表
-- ============================================================

CREATE TABLE IF NOT EXISTS `monitor_slow_request` (
    `id`          BIGINT       NOT NULL AUTO_INCREMENT,
    `request_id`  VARCHAR(64)  NOT NULL DEFAULT '',
    `method`      VARCHAR(8)   NOT NULL,
    `path`        VARCHAR(256) NOT NULL,
    `status_code` INT          NOT NULL,
    `duration_ms` INT          NOT NULL,
    `user_id`     BIGINT       NOT NULL DEFAULT 0,
    `created_at`  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_path` (`path`),
    KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='慢请求记录表';

CREATE TABLE IF NOT EXISTS `monitor_error_event` (
    `id`          BIGINT       NOT NULL AUTO_INCREMENT,
    `source`      VARCHAR(32)  NOT NULL COMMENT 'panic/mysql/redis/milvus/ai',
    `path`        VARCHAR(256) NOT NULL DEFAULT '',
    `request_id`  VARCHAR(64)  NOT NULL DEFAULT '',
    `message`     TEXT         NOT NULL,
    `count`       INT          NOT NULL DEFAULT 1 COMMENT '聚合计数',
    `created_at`  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_source` (`source`),
    KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='错误事件聚合表';

CREATE TABLE IF NOT EXISTS `alert_event` (
    `id`            BIGINT       NOT NULL COMMENT '雪花ID',
    `fingerprint`   VARCHAR(64)  NOT NULL COMMENT 'hash(alert_name+service+endpoint+severity)',
    `alert_name`    VARCHAR(128) NOT NULL,
    `service`       VARCHAR(64)  NOT NULL DEFAULT '',
    `endpoint`      VARCHAR(256) NOT NULL DEFAULT '',
    `severity`      VARCHAR(16)  NOT NULL DEFAULT 'warning',
    `status`        VARCHAR(16)  NOT NULL DEFAULT 'firing' COMMENT 'firing/acknowledged/resolved',
    `labels`        JSON         NOT NULL,
    `annotations`   JSON         NOT NULL,
    `repeat_count`  INT          NOT NULL DEFAULT 1,
    `first_seen_at` DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `last_seen_at`  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `resolved_at`   DATETIME     NULL,
    `created_at`    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_fingerprint` (`fingerprint`),
    KEY `idx_status` (`status`),
    KEY `idx_alert_name` (`alert_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='告警事件表';

CREATE TABLE IF NOT EXISTS `alert_enrichment` (
    `id`         BIGINT   NOT NULL AUTO_INCREMENT,
    `alert_id`   BIGINT   NOT NULL,
    `context`    JSON     NOT NULL COMMENT '告警增强上下文JSON',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_alert_id` (`alert_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='告警增强上下文表';

CREATE TABLE IF NOT EXISTS `incident` (
    `id`          BIGINT       NOT NULL COMMENT '雪花ID',
    `title`       VARCHAR(256) NOT NULL,
    `status`      VARCHAR(16)  NOT NULL DEFAULT 'open' COMMENT 'open/resolved',
    `severity`    VARCHAR(16)  NOT NULL DEFAULT 'warning',
    `created_at`  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='故障事件表';

CREATE TABLE IF NOT EXISTS `incident_alert` (
    `id`          BIGINT   NOT NULL AUTO_INCREMENT,
    `incident_id` BIGINT   NOT NULL,
    `alert_id`    BIGINT   NOT NULL,
    `created_at`  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_incident_alert` (`incident_id`, `alert_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='故障-告警关联表';

CREATE TABLE IF NOT EXISTS `deployment_change` (
    `id`          BIGINT       NOT NULL COMMENT '雪花ID',
    `service`     VARCHAR(64)  NOT NULL,
    `namespace`   VARCHAR(64)  NOT NULL DEFAULT 'default',
    `image`       VARCHAR(256) NOT NULL DEFAULT '',
    `old_image`   VARCHAR(256) NOT NULL DEFAULT '',
    `replicas`    INT          NOT NULL DEFAULT 0,
    `changed_by`  VARCHAR(64)  NOT NULL DEFAULT '',
    `source`      VARCHAR(16)  NOT NULL DEFAULT 'manual' COMMENT 'helm/kubectl/argocd/manual',
    `changed_at`  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `created_at`  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_service` (`service`),
    KEY `idx_changed_at` (`changed_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='部署变更记录表';

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

SET FOREIGN_KEY_CHECKS = 1;
