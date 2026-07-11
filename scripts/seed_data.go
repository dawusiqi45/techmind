//go:build ignore
// +build ignore

// TechMind 演示数据种子脚本
// 使用方式：go run scripts/seed_data.go
// 依赖：项目已编译并连接数据库

package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// ========== 模型定义 ==========

type User struct {
	ID        int64     `gorm:"column:id;primaryKey"`
	Username  string    `gorm:"column:username;uniqueIndex"`
	Password  string    `gorm:"column:password"`
	Email     string    `gorm:"column:email;uniqueIndex"`
	Avatar    string    `gorm:"column:avatar"`
	Role      int8      `gorm:"column:role"`
	Status    int8      `gorm:"column:status"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

type Article struct {
	ID            int64     `gorm:"column:id;primaryKey"`
	AuthorID      int64     `gorm:"column:author_id;index"`
	Title         string    `gorm:"column:title"`
	Content       string    `gorm:"column:content"`
	Summary       string    `gorm:"column:summary"`
	Cover         string    `gorm:"column:cover"`
	Status        int8      `gorm:"column:status"`
	IndexStatus   int8      `gorm:"column:index_status"`
	ViewCount     int       `gorm:"column:view_count"`
	LikeCount     int       `gorm:"column:like_count"`
	FavoriteCount int       `gorm:"column:favorite_count"`
	CommentCount  int       `gorm:"column:comment_count"`
	CreatedAt     time.Time `gorm:"column:created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at"`
}

type Tag struct {
	ID        int64     `gorm:"column:id;primaryKey"`
	Name      string    `gorm:"column:name;uniqueIndex"`
	HotScore  float64   `gorm:"column:hot_score"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

type ArticleTag struct {
	ID        int64     `gorm:"column:id;primaryKey;autoIncrement"`
	ArticleID int64     `gorm:"column:article_id"`
	TagID     int64     `gorm:"column:tag_id"`
	Source    string    `gorm:"column:source"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

type Comment struct {
	ID        int64     `gorm:"column:id;primaryKey"`
	ArticleID int64     `gorm:"column:article_id;index"`
	AuthorID  int64     `gorm:"column:author_id"`
	ParentID  int64     `gorm:"column:parent_id"`
	Content   string    `gorm:"column:content"`
	Status    int8      `gorm:"column:status"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

type Favorite struct {
	ID        int64     `gorm:"column:id;primaryKey;autoIncrement"`
	UserID    int64     `gorm:"column:user_id"`
	ArticleID int64     `gorm:"column:article_id"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

type Runbook struct {
	ID        int64     `gorm:"column:id;primaryKey"`
	Title     string    `gorm:"column:title"`
	Content   string    `gorm:"column:content"`
	AlertName string    `gorm:"column:alert_name"`
	Service   string    `gorm:"column:service"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

// ========== 演示数据 ==========

const demoAdminPassword = "TechMind123!"

var demoUsers = []User{
	{ID: 10001, Username: "alice", Password: demoPasswordHash(), Email: "alice@techmind.io", Avatar: "https://api.dicebear.com/7.x/avataaars/svg?seed=alice", Role: 0, Status: 1, CreatedAt: time.Now().Add(-30 * 24 * time.Hour), UpdatedAt: time.Now()},
	{ID: 10002, Username: "bob", Password: demoPasswordHash(), Email: "bob@techmind.io", Avatar: "https://api.dicebear.com/7.x/avataaars/svg?seed=bob", Role: 0, Status: 1, CreatedAt: time.Now().Add(-25 * 24 * time.Hour), UpdatedAt: time.Now()},
	{ID: 10003, Username: "admin", Password: demoPasswordHash(), Email: "admin@techmind.io", Avatar: "https://api.dicebear.com/7.x/avataaars/svg?seed=admin", Role: 1, Status: 1, CreatedAt: time.Now().Add(-60 * 24 * time.Hour), UpdatedAt: time.Now()},
}

func demoPasswordHash() string {
	hash, err := bcrypt.GenerateFromPassword([]byte(demoAdminPassword), bcrypt.DefaultCost)
	if err != nil {
		panic(fmt.Sprintf("hash demo password: %v", err))
	}
	return string(hash)
}

var demoTags = []Tag{
	{ID: 20001, Name: "Go", HotScore: 95.5, CreatedAt: time.Now().Add(-60 * 24 * time.Hour), UpdatedAt: time.Now()},
	{ID: 20002, Name: "Kubernetes", HotScore: 88.2, CreatedAt: time.Now().Add(-55 * 24 * time.Hour), UpdatedAt: time.Now()},
	{ID: 20003, Name: "Prometheus", HotScore: 76.8, CreatedAt: time.Now().Add(-50 * 24 * time.Hour), UpdatedAt: time.Now()},
	{ID: 20004, Name: "AI", HotScore: 92.1, CreatedAt: time.Now().Add(-45 * 24 * time.Hour), UpdatedAt: time.Now()},
	{ID: 20005, Name: "Gin", HotScore: 65.3, CreatedAt: time.Now().Add(-40 * 24 * time.Hour), UpdatedAt: time.Now()},
	{ID: 20006, Name: "Milvus", HotScore: 58.7, CreatedAt: time.Now().Add(-35 * 24 * time.Hour), UpdatedAt: time.Now()},
	{ID: 20007, Name: "SRE", HotScore: 82.4, CreatedAt: time.Now().Add(-30 * 24 * time.Hour), UpdatedAt: time.Now()},
	{ID: 20008, Name: "RAG", HotScore: 71.9, CreatedAt: time.Now().Add(-28 * 24 * time.Hour), UpdatedAt: time.Now()},
}

var demoArticles = []Article{
	{
		ID: 30001, AuthorID: 10001, Title: "TechMind 架构设计：从技术论坛到云原生智能可观测平台",
		Content: `# TechMind 架构设计

TechMind 是一个基于 Go/Gin 的技术论坛与云原生智能可观测平台。本文将详细介绍其架构设计。

## 整体架构

TechMind 采用分层架构：
- **Controller**：处理 HTTP 协议层
- **Logic**：业务编排层
- **DAO**：数据访问层（MySQL、Redis、Milvus）

## 技术栈

- Go 1.24 + Gin
- MySQL 8 + GORM
- Redis 7（缓存 + Stream）
- Milvus（向量检索）
- Prometheus + Alertmanager

## 核心特性

1. 技术论坛：文章、评论、标签、热榜
2. AI 增强：文章摘要、AI 标签、搜索总结
3. 可观测：指标采集、慢请求、错误事件
4. 告警中心：Alertmanager 接入、去重、增强
5. SRE Agent：MCP 工具 + RAG 诊断`,
		Summary: "TechMind 采用分层架构，整合技术论坛与云原生可观测能力。",
		Cover:   "https://picsum.photos/seed/techmind-arch/800/400",
		Status:  1, IndexStatus: 1, ViewCount: 1200, LikeCount: 85, FavoriteCount: 42, CommentCount: 15,
		CreatedAt: time.Now().Add(-10 * 24 * time.Hour), UpdatedAt: time.Now(),
	},
	{
		ID: 30002, AuthorID: 10002, Title: "Prometheus 指标设计实践：为 TechMind 埋点",
		Content: `# Prometheus 指标设计实践

在 TechMind 项目中，我们通过自定义 Prometheus 指标实现了完整的可观测能力。

## 核心指标

### HTTP 指标
- http_requests_total：请求总数
- http_request_duration_seconds：请求耗时直方图
- http_errors_total：错误请求数

### 慢请求检测
当请求耗时超过阈值时，自动写入 monitor_slow_request 表。

### AI 调用观测
- ai_calls_total：AI 调用总数
- ai_call_duration_seconds：AI 调用耗时
- ai_call_errors_total：AI 调用失败数
- ai_token_usage_total：Token 使用量

## 最佳实践

1. 指标命名遵循 Prometheus 最佳实践
2. 使用 Label 区分不同端点和状态码
3. 关键路径埋点不阻塞主流程`,
		Summary: "TechMind 通过自定义 Prometheus 指标实现完整可观测能力。",
		Cover:   "https://picsum.photos/seed/prometheus-metrics/800/400",
		Status:  1, IndexStatus: 1, ViewCount: 980, LikeCount: 62, FavoriteCount: 28, CommentCount: 8,
		CreatedAt: time.Now().Add(-8 * 24 * time.Hour), UpdatedAt: time.Now(),
	},
	{
		ID: 30003, AuthorID: 10001, Title: "Redis Stream 在异步任务队列中的应用",
		Content: `# Redis Stream 异步任务队列

TechMind 使用 Redis Stream 作为异步任务队列，处理 AI 摘要、标签生成和向量索引。

## 为什么选择 Redis Stream

1. **持久化**：消息不丢失
2. **Consumer Group**：支持多消费者竞争消费
3. **ACK 机制**：消息确认后删除，支持重试
4. **Pending 查询**：可查看未确认消息

## 实现细节

    // 写入任务
    stream.EnqueueAITask(ctx, "article.summary", articleID)

    // Worker 消费
    for {
        msgs, err := stream.ReadAITasks(ctx, group, consumer, count)
        // 处理任务...
        stream.AckAITask(ctx, group, msg.ID)
    }

## 重试和死信

- 失败任务自动重试，最多 3 次
- 超过重试次数进入死信队列
- 死信队列可通过后台查询和处理`,
		Summary: "TechMind 使用 Redis Stream 实现可靠的异步任务队列。",
		Cover:   "https://picsum.photos/seed/redis-stream/800/400",
		Status:  1, IndexStatus: 1, ViewCount: 750, LikeCount: 45, FavoriteCount: 19, CommentCount: 6,
		CreatedAt: time.Now().Add(-6 * 24 * time.Hour), UpdatedAt: time.Now(),
	},
	{
		ID: 30004, AuthorID: 10003, Title: "SRE Agent 诊断报告设计：从告警到根因分析",
		Content: `# SRE Agent 诊断报告设计

TechMind 的 SRE Agent 参考 HolmesGPT，但只做项目内可控版本。

## 诊断流程

1. **触发**：管理员点击诊断 / 告警触发诊断
2. **证据采集**：调用 MCP 只读工具采集指标、日志、队列等证据
3. **RAG 检索**：从 Milvus 检索相关 Runbook 和历史报告
4. **变更关联**：查询最近部署变更
5. **LLM 汇总**：生成结构化诊断报告

## 诊断报告结构

    {
      "summary": "搜索接口最近10分钟P95升高",
      "evidence": ["Prometheus显示 P95 = 1.6s"],
      "root_cause": "Milvus检索延迟升高叠加AI搜索总结耗时",
      "impact": "影响搜索页响应速度",
      "suggestions": [
        "对搜索总结结果增加Redis缓存",
        "AI总结超时时降级为普通搜索"
      ]
    }

## MCP Toolset

- prometheus_query：查询指标
- alert_query：查询告警
- slow_request_query：查询慢请求
- k8s_resource_query：查询 Pod/Deployment`,
		Summary: "TechMind SRE Agent 通过 MCP 工具 + RAG 生成结构化诊断报告。",
		Cover:   "https://picsum.photos/seed/sre-agent/800/400",
		Status:  1, IndexStatus: 1, ViewCount: 1100, LikeCount: 78, FavoriteCount: 35, CommentCount: 12,
		CreatedAt: time.Now().Add(-5 * 24 * time.Hour), UpdatedAt: time.Now(),
	},
	{
		ID: 30005, AuthorID: 10002, Title: "Milvus 向量检索在语义搜索中的实践",
		Content: `# Milvus 向量检索实践

TechMind 使用 Milvus 实现文章语义搜索，提升搜索体验。

## 向量索引流程

1. 文章发布时，切分 Markdown 内容
2. 调用 Embedding 模型生成向量
3. 写入 Milvus 并关联 article_id
4. 同时保存 chunk 正文到 article_chunk 表

## 搜索流程

1. 用户输入搜索关键词
2. MySQL 关键词搜索（标题、内容 LIKE 查询）
3. Milvus 语义搜索（向量相似度）
4. 结果融合排序
5. SearchSummarySkill 生成搜索总结

## 性能优化

- 按 query hash 缓存搜索结果
- 超时降级为纯关键词搜索
- Milvus 连接池管理`,
		Summary: "TechMind 使用 Milvus 实现语义搜索，提升文章检索效果。",
		Cover:   "https://picsum.photos/seed/milvus-search/800/400",
		Status:  1, IndexStatus: 1, ViewCount: 850, LikeCount: 52, FavoriteCount: 22, CommentCount: 7,
		CreatedAt: time.Now().Add(-3 * 24 * time.Hour), UpdatedAt: time.Now(),
	},
}

var demoArticleTags = []ArticleTag{
	{ID: 40001, ArticleID: 30001, TagID: 20001, Source: "manual", CreatedAt: time.Now()},
	{ID: 40002, ArticleID: 30001, TagID: 20002, Source: "ai", CreatedAt: time.Now()},
	{ID: 40003, ArticleID: 30001, TagID: 20007, Source: "ai", CreatedAt: time.Now()},
	{ID: 40004, ArticleID: 30002, TagID: 20003, Source: "manual", CreatedAt: time.Now()},
	{ID: 40005, ArticleID: 30002, TagID: 20007, Source: "ai", CreatedAt: time.Now()},
	{ID: 40006, ArticleID: 30003, TagID: 20001, Source: "manual", CreatedAt: time.Now()},
	{ID: 40007, ArticleID: 30004, TagID: 20007, Source: "manual", CreatedAt: time.Now()},
	{ID: 40008, ArticleID: 30004, TagID: 20004, Source: "ai", CreatedAt: time.Now()},
	{ID: 40009, ArticleID: 30005, TagID: 20006, Source: "manual", CreatedAt: time.Now()},
	{ID: 40010, ArticleID: 30005, TagID: 20004, Source: "ai", CreatedAt: time.Now()},
}

var demoComments = []Comment{
	{ID: 50001, ArticleID: 30001, AuthorID: 10002, ParentID: 0, Content: "架构设计非常清晰，期待更多细节！", Status: 1, CreatedAt: time.Now().Add(-9 * 24 * time.Hour), UpdatedAt: time.Now()},
	{ID: 50002, ArticleID: 30001, AuthorID: 10001, ParentID: 50001, Content: "感谢支持，后续会出系列文章深入讲解。", Status: 1, CreatedAt: time.Now().Add(-8 * 24 * time.Hour), UpdatedAt: time.Now()},
	{ID: 50003, ArticleID: 30002, AuthorID: 10001, ParentID: 0, Content: "Prometheus 指标设计很详细，学习了。", Status: 1, CreatedAt: time.Now().Add(-7 * 24 * time.Hour), UpdatedAt: time.Now()},
	{ID: 50004, ArticleID: 30004, AuthorID: 10001, ParentID: 0, Content: "SRE Agent 的设计思路很棒，特别是 MCP 工具封装。", Status: 1, CreatedAt: time.Now().Add(-4 * 24 * time.Hour), UpdatedAt: time.Now()},
	{ID: 50005, ArticleID: 30005, AuthorID: 10003, ParentID: 0, Content: "Milvus 向量检索在搜索中的应用很有价值。", Status: 1, CreatedAt: time.Now().Add(-2 * 24 * time.Hour), UpdatedAt: time.Now()},
}

var demoFavorites = []Favorite{
	{ID: 60001, UserID: 10002, ArticleID: 30001, CreatedAt: time.Now().Add(-8 * 24 * time.Hour)},
	{ID: 60002, UserID: 10003, ArticleID: 30001, CreatedAt: time.Now().Add(-7 * 24 * time.Hour)},
	{ID: 60003, UserID: 10001, ArticleID: 30002, CreatedAt: time.Now().Add(-6 * 24 * time.Hour)},
	{ID: 60004, UserID: 10002, ArticleID: 30004, CreatedAt: time.Now().Add(-3 * 24 * time.Hour)},
	{ID: 60005, UserID: 10001, ArticleID: 30005, CreatedAt: time.Now().Add(-1 * 24 * time.Hour)},
}

var demoRunbooks = []Runbook{
	{
		ID: 70001, Title: "APIHighErrorRate 告警处理手册",
		Content: `# APIHighErrorRate 告警处理

## 告警说明
API 5xx 错误率超过阈值，可能原因包括：
- 数据库连接异常
- Redis 连接异常
- 依赖服务（Milvus）不可用

## 排查步骤

1. 查看错误事件聚合表，定位主要错误类型
2. 检查数据库连接池状态
3. 检查 Redis 连接状态
4. 检查 Milvus 服务状态
5. 查看最近部署变更

## 应急措施

- 如果是数据库问题，检查连接池配置
- 如果是 Redis 问题，检查内存和连接数
- 如果是 Milvus 问题，检查 collection 状态`,
		AlertName: "APIHighErrorRate", Service: "techmind-server", CreatedAt: time.Now(), UpdatedAt: time.Now(),
	},
	{
		ID: 70002, Title: "RedisStreamBacklogHigh 告警处理手册",
		Content: `# RedisStreamBacklogHigh 告警处理

## 告警说明
AI 或诊断任务队列积压，可能原因包括：
- Worker 实例不足
- Worker 消费异常
- AI 调用超时

## 排查步骤

1. 检查 Worker 实例数量和状态
2. 查看 pending 消息数量
3. 检查 Worker 消费延迟
4. 查看失败任务和死信队列

## 应急措施

- 增加 Worker 实例
- 重启异常 Worker
- 清理死信队列`,
		AlertName: "RedisStreamBacklogHigh", Service: "techmind-worker", CreatedAt: time.Now(), UpdatedAt: time.Now(),
	},
	{
		ID: 70003, Title: "AICallFailureHigh 告警处理手册",
		Content: `# AICallFailureHigh 告警处理

## 告警说明
AI 调用失败率超过阈值，可能原因包括：
- LLM 服务不可用
- API Key 过期
- 请求超时
- 网络问题

## 排查步骤

1. 检查 LLM 服务健康状态
2. 验证 API Key 有效性
3. 检查请求日志中的错误信息
4. 查看 AI 调用记录表

## 应急措施

- 检查 API Key 配置
- 增加超时时间
- 启用降级策略`,
		AlertName: "AICallFailureHigh", Service: "techmind-server", CreatedAt: time.Now(), UpdatedAt: time.Now(),
	},
}

// ========== 种子逻辑 ==========

func main() {
	// 默认连接参数，可通过环境变量覆盖
	dsn := getEnv("MYSQL_DSN", "techmind:techmind@tcp(127.0.0.1:3306)/techmind?parseTime=true&charset=utf8mb4&loc=Local")

	fmt.Println("=== TechMind 演示数据种子脚本 ===")
	fmt.Printf("连接 MySQL: %s\n", maskDSN(dsn))

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("连接 MySQL 失败: %v", err)
	}

	fmt.Println("MySQL 连接成功，开始写入演示数据...")

	// 清空已有演示数据（避免重复）
	cleanDemoData(db)

	// 写入用户
	if err := seedUsers(db); err != nil {
		log.Fatalf("写入用户失败: %v", err)
	}
	fmt.Println("✅ 用户数据写入完成 (3)")

	// 写入标签
	if err := seedTags(db); err != nil {
		log.Fatalf("写入标签失败: %v", err)
	}
	fmt.Println("✅ 标签数据写入完成 (8)")

	// 写入文章
	if err := seedArticles(db); err != nil {
		log.Fatalf("写入文章失败: %v", err)
	}
	fmt.Println("✅ 文章数据写入完成 (5)")

	// 写入文章标签关联
	if err := seedArticleTags(db); err != nil {
		log.Fatalf("写入文章标签关联失败: %v", err)
	}
	fmt.Println("✅ 文章标签关联写入完成 (10)")

	// 写入评论
	if err := seedComments(db); err != nil {
		log.Fatalf("写入评论失败: %v", err)
	}
	fmt.Println("✅ 评论数据写入完成 (5)")

	// 写入收藏
	if err := seedFavorites(db); err != nil {
		log.Fatalf("写入收藏失败: %v", err)
	}
	fmt.Println("✅ 收藏数据写入完成 (5)")

	// 写入 Runbook
	if err := seedRunbooks(db); err != nil {
		log.Fatalf("写入 Runbook 失败: %v", err)
	}
	fmt.Println("✅ Runbook 数据写入完成 (3)")

	fmt.Println("\n🎉 演示数据种子完成！")
	fmt.Printf("提示：演示管理员账号为 admin，密码为 %s；仅限本地演示，生产环境必须改用独立管理员和强密码。\n", demoAdminPassword)
}

func cleanDemoData(db *gorm.DB) {
	// 删除演示数据范围内的记录
	db.Where("id IN (?)", []int64{10001, 10002, 10003}).Delete(&User{})
	db.Where("id IN (?)", []int64{20001, 20002, 20003, 20004, 20005, 20006, 20007, 20008}).Delete(&Tag{})
	db.Where("id IN (?)", []int64{30001, 30002, 30003, 30004, 30005}).Delete(&Article{})
	db.Where("id <= ?", 40010).Delete(&ArticleTag{})
	db.Where("id IN (?)", []int64{50001, 50002, 50003, 50004, 50005}).Delete(&Comment{})
	db.Where("id <= ?", 60005).Delete(&Favorite{})
	db.Where("id IN (?)", []int64{70001, 70002, 70003}).Delete(&Runbook{})
}

func seedUsers(db *gorm.DB) error {
	return db.Create(&demoUsers).Error
}

func seedTags(db *gorm.DB) error {
	return db.Create(&demoTags).Error
}

func seedArticles(db *gorm.DB) error {
	return db.Create(&demoArticles).Error
}

func seedArticleTags(db *gorm.DB) error {
	return db.Create(&demoArticleTags).Error
}

func seedComments(db *gorm.DB) error {
	return db.Create(&demoComments).Error
}

func seedFavorites(db *gorm.DB) error {
	return db.Create(&demoFavorites).Error
}

func seedRunbooks(db *gorm.DB) error {
	return db.Create(&demoRunbooks).Error
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func maskDSN(dsn string) string {
	// 简单脱敏处理
	if len(dsn) > 20 {
		return dsn[:20] + "..."
	}
	return dsn
}
