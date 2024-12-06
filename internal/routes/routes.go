package routes

import (
	"mywallet/internal/api"
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
	app_api := r.Group("api/app")
	{
		app_api.POST("/wallet/deposit", api.Deposit)
		app_api.POST("/wallet/withdraw", api.Withdraw)
		app_api.POST("/wallet/transfer", api.Transfer)
		app_api.GET("/wallet/balance/:address", api.GetBalance)
		app_api.GET("/wallet/transactions/:address", api.GetTransactions)
	}
}
