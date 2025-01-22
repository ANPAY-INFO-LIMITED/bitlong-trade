package routers

import (
	"github.com/gin-gonic/gin"
	"trade/handlers"
	"trade/middleware"
)

func SetupAssetInfoRouter(router *gin.Engine) *gin.Engine {
	assetInfo := router.Group("/asset_info")
	assetInfo.Use(middleware.AuthMiddleware())
	{
		assetInfo.POST("/get_name/slice", handlers.GetAssetNames)
		assetInfo.POST("/get_name/slice/map", handlers.GetAssetsNameMap)
	}
	return router
}
