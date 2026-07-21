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
	if err := r.SetTrustedProxies(settings.Conf.Server.TrustedProxies); err != nil {
		panic("invalid trusted proxy configuration: " + err.Error())
	}
	r.Use(middleware.CORS())
	r.Use(middleware.BodyLimit(4 << 20))
	r.Use(middleware.RequestID())
	r.Use(middleware.Logger())
	r.Use(middleware.Metrics())
	r.Use(middleware.SlowRequest(800 * time.Millisecond))
	r.Use(middleware.Recovery())

	r.Static("/uploads", "./uploads")

	// --- 系统路由 ---
	r.GET("/healthz", healthzHandler)
	r.GET("/readyz", readyzHandler)
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// --- API v1 ---
	v1 := r.Group("/api/v1")

	// Auth（无需鉴权）
	auth := v1.Group("/auth")
	auth.Use(middleware.RateLimit(&settings.Conf.RateLimit))
	{
		auth.POST("/register", controller.Register)
		auth.POST("/login", controller.Login)
		auth.POST("/refresh", controller.RefreshToken)
	}

	// 需要鉴权的路由
	authed := v1.Group("")
	authed.Use(middleware.RateLimit(&settings.Conf.RateLimit))
	authed.Use(middleware.JWT())
	authed.Use(middleware.RateLimit(&settings.Conf.RateLimit))
	{
		// 用户
		authed.GET("/user/profile", controller.GetProfile)
		authed.GET("/user/favorites", controller.ListUserFavorites)
		authed.GET("/user/likes", controller.ListUserLikes)
		authed.GET("/user/articles", controller.ListUserArticles)
		authed.PUT("/user/profile", controller.UpdateProfile)
		authed.POST("/user/avatar", controller.UploadAvatar)

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
	public.Use(middleware.RateLimit(&settings.Conf.RateLimit))
	{
		public.GET("/articles", controller.ListArticles)
		public.GET("/articles/hot", controller.GetHotArticles)
		public.GET("/articles/:id", controller.GetArticle)
		public.GET("/articles/:id/comments", controller.ListComments)
		public.GET("/tags", controller.ListTags)
		public.GET("/search", controller.SearchArticles)
	}

	// Alertmanager 无法携带用户 JWT，使用独立的 Bearer Webhook 令牌鉴权。
	webhook := v1.Group("/alerts")
	webhook.Use(middleware.RateLimit(&settings.Conf.RateLimit))
	{
		webhook.POST("/webhook", controller.AlertWebhook)
	}

	// 管理后台路由（需 JWT 和服务端管理员角色校验）
	admin := v1.Group("/admin")
	admin.Use(middleware.RateLimit(&settings.Conf.RateLimit))
	admin.Use(middleware.JWT())
	admin.Use(middleware.RateLimit(&settings.Conf.RateLimit))
	admin.Use(middleware.RequireAdmin())
	{
		// 监控后台
		admin.GET("/monitor/overview", controller.MonitorOverview)
		admin.GET("/monitor/slow-requests", controller.MonitorSlowRequests)
		admin.GET("/monitor/errors", controller.MonitorErrors)
		admin.GET("/monitor/queues", controller.MonitorQueues)
		admin.GET("/monitor/ai-calls", controller.MonitorAICalls)

		// 告警中心
		admin.GET("/alerts", controller.ListAlerts)
		admin.GET("/alerts/:id", controller.GetAlertDetail)
		admin.POST("/alerts/:id/ack", controller.AcknowledgeAlert)
		admin.POST("/alerts/:id/diagnose", controller.AlertDiagnose)

		// SRE 诊断报告
		admin.POST("/ops/diagnose", controller.ManualDiagnose)
		admin.GET("/ops/reports", controller.ListOpsReports)
		admin.GET("/ops/reports/:id", controller.GetOpsReport)
		admin.GET("/ops/reports/:id/timeline", controller.GetOpsReportTimeline)

		// 故障事件：由告警诊断自动聚合，管理员可查看并手动关闭。
		admin.GET("/incidents", controller.ListIncidents)
		admin.GET("/incidents/:id", controller.GetIncident)
		admin.POST("/incidents/:id/resolve", controller.ResolveIncident)

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
