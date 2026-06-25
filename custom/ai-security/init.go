package ai_security

import (
	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/custom/ai-security/api"
	aisecmiddleware "github.com/QuantumNous/new-api/custom/ai-security/middleware"
	"github.com/QuantumNous/new-api/custom/ai-security/migration"
	"github.com/QuantumNous/new-api/custom/ai-security/model"
	"github.com/QuantumNous/new-api/custom/ai-security/seed"
	"github.com/QuantumNous/new-api/custom/ai-security/service"

	"github.com/gin-gonic/gin"
)

// Init 初始化 ai-security 模块
func Init() error {
	common.SysLog("[ai-security] initializing module")

	// 1. 执行数据库迁移
	if err := migration.Run(); err != nil {
		return err
	}

	// 2. 初始化默认配置
	if err := model.InitDefaultConfigs(); err != nil {
		return err
	}

	// 3. 初始化默认规则
	if err := seed.InitDefaultRules(); err != nil {
		return err
	}

	// 4. 初始化缓存
	service.InitCache()

	// 5. 启动每日统计聚合调度器
	service.StartDailyStatsScheduler()

	common.SysLog("[ai-security] module initialized")
	return nil
}

// RegisterRoutes 注册后端路由
func RegisterRoutes(router *gin.RouterGroup) {
	securityRoute := router.Group("/ai-security")
	{
		securityRoute.GET("/configs", api.GetConfigs)
		securityRoute.PUT("/configs", api.UpdateConfigs)
		securityRoute.GET("/status", api.GetStatus)
		securityRoute.POST("/install", api.Install)

		securityRoute.GET("/groups", api.ListGroups)
		securityRoute.POST("/groups", api.CreateGroup)
		securityRoute.PUT("/groups/:id", api.UpdateGroup)
		securityRoute.PATCH("/groups/:id/status", api.UpdateGroupStatus)
		securityRoute.DELETE("/groups/:id", api.DeleteGroup)
		securityRoute.POST("/groups/:id/copy", api.CopyGroup)

		securityRoute.GET("/rules", api.ListRules)
		securityRoute.POST("/rules", api.CreateRule)
		securityRoute.PUT("/rules/:id", api.UpdateRule)
		securityRoute.DELETE("/rules/:id", api.DeleteRule)
		securityRoute.POST("/rules/:id/test", api.TestRule)
		securityRoute.PATCH("/rules/:id/status", api.UpdateRuleStatus)
		securityRoute.GET("/policies", api.ListPolicies)
		securityRoute.POST("/policies", api.CreatePolicy)
		securityRoute.PUT("/policies/:id", api.UpdatePolicy)
		securityRoute.DELETE("/policies/:id", api.DeletePolicy)

		securityRoute.GET("/logs", api.ListHitLogs)
		securityRoute.GET("/logs/export", api.ExportHitLogs)
		securityRoute.GET("/dashboard", api.GetDashboard)
		securityRoute.GET("/dashboard/trend", api.GetDailyTrend)

		securityRoute.POST("/sync/official-sensitive-words", api.SyncOfficialSensitiveWords)
	}
}

// RegisterRelayMiddleware 注册 relay 中间件
func RegisterRelayMiddleware(router *gin.RouterGroup) {
	// 在 relay-router.go 中调用
	// router.Use(ai_security.CheckRequest())
	// router.Use(ai_security.CheckResponse())
}

// CheckRequest 请求检测中间件
func CheckRequest() gin.HandlerFunc {
	return aisecmiddleware.CheckRequest()
}

// CheckResponse 响应检测中间件
func CheckResponse() gin.HandlerFunc {
	return aisecmiddleware.CheckResponse()
}
