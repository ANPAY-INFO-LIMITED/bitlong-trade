package routers

import (
	"github.com/gin-gonic/gin"
	"trade/config"
	"trade/handlers"
)

func SetupAssetBalanceBackendRouter(router *gin.Engine) *gin.Engine {

	//assetBalanceBackend := router.Group("/asset_balance_backend")

	username := config.GetLoadConfig().AdminUser.Username
	password := config.GetLoadConfig().AdminUser.Password

	authorized := router.Group("/asset_balance_backend", gin.BasicAuth(gin.Accounts{
		username: password,
	}))

	assetBalanceInfo := authorized.Group("/asset_balance_info")
	{
		assetBalanceInfo.GET("/get/limit_offset", handlers.GetAssetBalanceLimitAndOffset)
		assetBalanceInfo.GET("/get/count", handlers.GetAssetBalanceCount)
		assetBalanceInfo.GET("/get/username", handlers.QueryAssetBalanceInfoByUsername)
		assetBalanceInfo.GET("/query/all_asset_ids", handlers.QueryAllAssetBalanceAssetIds)
	}

	AssetBalanceHistoryInfo := authorized.Group("/asset_balance_history_info")
	{
		AssetBalanceHistoryInfo.GET("/get/limit_offset", handlers.GetAssetBalanceHistoryLimitAndOffset)
		AssetBalanceHistoryInfo.GET("/get/count", handlers.GetAssetBalanceHistoryCount)
		AssetBalanceHistoryInfo.GET("/get/username", handlers.QueryAssetBalanceHistoryInfoByUsername)
		AssetBalanceHistoryInfo.GET("/query/all_asset_ids", handlers.QueryAllAssetBalanceHistoryAssetIds)
	}

	return router
}
