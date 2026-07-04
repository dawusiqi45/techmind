package router

import (
	"net/http"
	"time"

	"techmind/internal/controller"
	"techmind/internal/middleware"
	"techmind/internal/pkg/health"
	"techmind/internal/pkg/settings"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Setup 初始化 Gin 引擎并注册所有路由，返回 *gin.Engine
func Setup(mode string) *gin.Engine {
	if mode != "local" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(middleware.CORS())
	r.Use(middleware.RequestID())
	r.Use(middleware.Logger())
	r.Use(middleware.Metrics())
	r.Use(middleware.RateLimit(&settings.Conf.RateLimit))
	r.Use(middleware.SlowRequest(800 * time.Millisecond))
	r.Use(middleware.Recovery())

	// --- 系统路由 ---
	r.GET("/healthz", healthzHandler)
	r.GET("/readyz", readyzHandler)
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// --- API v1 ---
	v1 := r.Group("/api/v1")

	// Auth（无需鉴权）
	auth := v1.Group("/auth")
	{
		auth.POST("/register", controller.Register)
		auth.POST("/login", controller.Login)
		auth.POST("/refresh", controller.RefreshToken)
	}

	// 需要鉴权的路由
	authed := v1.Group("")
	authed.Use(middleware.JWT())
	{
		// 用户
		authed.GET("/user/profile", controller.GetProfile)

		// 文章
		authed.POST("/articles", controller.CreateArticle)
		authed.PUT("/articles/:id", controller.UpdateArticle)
		authed.DELETE("/articles/:id", controller.DeleteArticle)
		authed.POST("/articles/:id/like", controller.LikeArticle)
		authed.POST("/articles/:id/favorite", controller.FavoriteArticle)
		authed.POST("/articles/:id/comments", controller.CreateComment)
	}

	// 公开路由（无需鉴权）
	public := v1.Group("")
	{
		public.GET("/articles", controller.ListArticles)
		public.GET("/articles/hot", controller.GetHotArticles)
		public.GET("/articles/:id", controller.GetArticle)
		public.GET("/articles/:id/comments", controller.ListComments)
		public.GET("/tags", controller.ListTags)
		public.GET("/search", controller.SearchArticles)
	}

	// 管理后台路由（需鉴权）
	admin := v1.Group("/admin")
	admin.Use(middleware.JWT())
	{
		// 监控后台
		admin.GET("/monitor/overview", controller.MonitorOverview)
		admin.GET("/monitor/slow-requests", controller.MonitorSlowRequests)
		admin.GET("/monitor/errors", controller.MonitorErrors)
		admin.GET("/monitor/queues", controller.MonitorQueues)
		admin.GET("/monitor/ai-calls", controller.MonitorAICalls)

		// 告警中心
		admin.POST("/alerts/webhook", controller.AlertWebhook)
		admin.GET("/alerts", controller.ListAlerts)
		admin.GET("/alerts/:id", controller.GetAlertDetail)
		admin.POST("/alerts/:id/ack", controller.AcknowledgeAlert)
		admin.POST("/alerts/:id/diagnose", controller.AlertDiagnose)

		// SRE 诊断报告
		admin.POST("/ops/diagnose", controller.ManualDiagnose)
		admin.GET("/ops/reports", controller.ListOpsReports)
		admin.GET("/ops/reports/:id", controller.GetOpsReport)

		// 部署变更
		admin.POST("/deployment-changes", controller.RecordDeploymentChange)
		admin.GET("/deployment-changes", controller.ListDeploymentChanges)

		// Runbook 管理
		admin.POST("/runbooks", controller.CreateRunbook)
		admin.GET("/runbooks", controller.ListRunbooks)
	}

	return r
}

func healthzHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func readyzHandler(c *gin.Context) {
	result := health.Readyz()
	status := http.StatusOK
	if !result.Ready {
		status = http.StatusServiceUnavailable
	}
	c.JSON(status, result)
}
