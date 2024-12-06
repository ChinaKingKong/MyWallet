package routes

import (
	"mywallet/internal/api"
	"mywallet/internal/config"
	"mywallet/pkg/logger"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

// 路由配置
func InitRouter() *gin.Engine {
	route := gin.Default()
	store := cookie.NewStore([]byte("mywallet-secret"))
	route.Use(sessions.Sessions("mywallet-session", store))
	route.StaticFS("/static", http.Dir("./static"))

	//App应用路由
	InitAppRouter(route)

	return route
}

// 初始化App应用路由
func InitAppRouter(r *gin.Engine) {
	app_api := r.Group("api/wallet")
	{
		cfg, _ := config.Load()
		logger := logger.NewLogger()
		server := api.NewServer(cfg, logger)
		if server != nil {
			app_api.POST("/deposit", server.Deposit)
			app_api.POST("/withdraw", server.Withdraw)
			app_api.POST("/transfer", server.Transfer)
			app_api.GET("/balance/:address", server.GetBalance)
			app_api.GET("/transactions/:address", server.GetTransactions)
		}
	}
}
