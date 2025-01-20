package routers

import (
	"github.com/gin-gonic/gin"
	"trade/config"
	"trade/handlers"
)

func SetupAssetMoreInfo(router *gin.Engine) *gin.Engine {
	username := config.GetLoadConfig().AdminUser.Username
	password := config.GetLoadConfig().AdminUser.Password

	authorized := router.Group("/asset_more_info", gin.BasicAuth(gin.Accounts{
		username: password,
	}))
	authorized.GET("/get/asset_balance_info_count", handlers.GetAssetBalanceInfoCount)
	authorized.GET("/get/asset_balance_info", handlers.GetAssetBalanceInfo)
	authorized.GET("/get/account_asset_transfer_count", handlers.GetAccountAssetTransferCount)
	authorized.GET("/get/account_asset_transfer", handlers.GetAccountAssetTransfer)
	authorized.GET("/get/asset_managed_utxo_info_count", handlers.GetAssetManagedUtxoInfoCount)
	authorized.GET("/get/asset_managed_utxo_info", handlers.GetAssetManagedUtxoInfo)
	authorized.GET("/get/asset_transfer_50", handlers.GetAssetTransferCombinedSliceByAssetIdLimit)
	return router
}
