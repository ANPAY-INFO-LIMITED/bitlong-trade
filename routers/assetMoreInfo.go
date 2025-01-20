package routers

import (
	"github.com/gin-gonic/gin"
	"trade/config"
	"trade/handlers"
)

func SetupAssetMoreInfo(router *gin.Engine) *gin.Engine {
	assetMoreInfo := router.Group("/asset_more_info")
	{
		assetMoreInfo.GET("/get/asset_balance_info/count", handlers.GetAssetBalanceInfoCount)
		assetMoreInfo.GET("/get/asset_balance_info", handlers.GetAssetBalanceInfo)
		assetMoreInfo.GET("/get/account_asset_transfer/count", handlers.GetAccountAssetTransferCount)
		assetMoreInfo.GET("/get/account_asset_transfer", handlers.GetAccountAssetTransfer)
		assetMoreInfo.GET("/get/asset_managed_utxo_info/count", handlers.GetAssetManagedUtxoInfoCount)
		assetMoreInfo.GET("/get/asset_managed_utxo_info", handlers.GetAssetManagedUtxoInfo)
		assetMoreInfo.GET("/get/asset_transfer_50", handlers.GetAssetTransferCombinedSliceByAssetIdLimit)
		assetMoreInfo.GET("/get/asset_burn_total", handlers.GetAssetBurnTotal)
	}

	username := config.GetLoadConfig().AdminUser.Username
	password := config.GetLoadConfig().AdminUser.Password

	authorized := router.Group("/asset_more_info/auth", gin.BasicAuth(gin.Accounts{
		username: password,
	}))
	authorized.GET("/get/asset_balance_info/count", handlers.GetAssetBalanceInfoCount)
	authorized.GET("/get/asset_balance_info", handlers.GetAssetBalanceInfo)
	authorized.GET("/get/account_asset_transfer/count", handlers.GetAccountAssetTransferCount)
	authorized.GET("/get/account_asset_transfer", handlers.GetAccountAssetTransfer)
	authorized.GET("/get/asset_managed_utxo_info/count", handlers.GetAssetManagedUtxoInfoCount)
	authorized.GET("/get/asset_managed_utxo_info", handlers.GetAssetManagedUtxoInfo)
	authorized.GET("/get/asset_transfer_50", handlers.GetAssetTransferCombinedSliceByAssetIdLimit)
	authorized.GET("/get/asset_burn_total", handlers.GetAssetBurnTotal)
	return router
}
