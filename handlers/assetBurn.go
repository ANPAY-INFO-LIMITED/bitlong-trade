package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"net/http"
	"trade/models"
	"trade/services"
	"trade/services/btldb"
)

func GetAssetBurnByUserId(c *gin.Context) {
	username := c.MustGet("username").(string)
	userId, err := services.NameToId(username)
	if err != nil {
		c.JSON(http.StatusOK, models.JsonResult{
			Success: false,
			Error:   err.Error(),
			Code:    models.NameToIdErr,
			Data:    nil,
		})
		return
	}
	assetBurns, err := services.GetAssetBurnsByUserId(userId)
	if err != nil {
		c.JSON(http.StatusOK, models.JsonResult{
			Success: false,
			Error:   err.Error(),
			Code:    models.GetAssetBurnsByUserIdErr,
			Data:    nil,
		})
		return
	}
	c.JSON(http.StatusOK, models.JsonResult{
		Success: true,
		Error:   models.SuccessErr,
		Code:    models.SUCCESS,
		Data:    assetBurns,
	})
}

func SetAssetBurn(c *gin.Context) {
	username := c.MustGet("username").(string)
	userId, err := services.NameToId(username)
	if err != nil {
		c.JSON(http.StatusOK, models.JsonResult{
			Success: false,
			Error:   err.Error(),
			Code:    models.NameToIdErr,
			Data:    nil,
		})
		return
	}
	var assetBurnSetRequest models.AssetBurnSetRequest
	err = c.ShouldBindJSON(&assetBurnSetRequest)
	if err != nil {
		c.JSON(http.StatusOK, models.JsonResult{
			Success: false,
			Error:   err.Error(),
			Code:    models.ShouldBindJsonErr,
			Data:    nil,
		})
		return
	}
	assetBurn := services.ProcessAssetBurnSetRequest(userId, username, &assetBurnSetRequest)
	err = btldb.CreateAssetBurn(assetBurn)
	if err != nil {
		c.JSON(http.StatusOK, models.JsonResult{
			Success: false,
			Error:   err.Error(),
			Code:    models.CreateAssetBurnErr,
			Data:    nil,
		})
		return
	}
	c.JSON(http.StatusOK, models.JsonResult{
		Success: true,
		Error:   models.SuccessErr,
		Code:    models.SUCCESS,
		Data:    nil,
	})
}

func GetAssetBurnByAssetId(c *gin.Context) {
	assetId := c.Param("asset_id")
	assetBurnTotal, err := services.GetAssetBurnTotal(assetId)
	if err != nil {
		c.JSON(http.StatusOK, models.JsonResult{
			Success: false,
			Error:   err.Error(),
			Code:    models.GetAssetBurnTotalErr,
			Data:    nil,
		})
		return
	}
	c.JSON(http.StatusOK, models.JsonResult{
		Success: true,
		Error:   models.SuccessErr,
		Code:    models.SUCCESS,
		Data:    assetBurnTotal.TotalAmount,
	})
}

func GetAllAssetBurnSimplified(c *gin.Context) {
	assetBurnSimplified, err := services.GetAllAssetBurnSimplified()
	if err != nil {
		c.JSON(http.StatusOK, models.JsonResult{
			Success: false,
			Error:   err.Error(),
			Code:    models.GetAllAssetBurnSimplifiedErr,
			Data:    nil,
		})
		return
	}
	c.JSON(http.StatusOK, models.JsonResult{
		Success: true,
		Error:   models.SuccessErr,
		Code:    models.SUCCESS,
		Data:    assetBurnSimplified,
	})
}

func GetAssetBurnAmount(c *gin.Context) {
	var req struct {
		AssetId string `json:"asset_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		logrus.Errorln(errors.Wrap(err, "c.ShouldBindJSON"))
		c.JSON(http.StatusOK, models.RespInt{
			Code: models.ToCode(models.ShouldBindJsonErr),
			Msg:  err.Error(),
			Data: 0,
		})
		return
	}

	if req.AssetId == "" {
		c.JSON(http.StatusOK, models.RespInt{
			Code: models.ToCode(models.InvalidReq),
			Msg:  invalidReq.Error(),
			Data: 0,
		})
		return
	}

	assetBurnTotal, err := services.GetAssetBurnTotal(req.AssetId)

	if err != nil {
		c.JSON(http.StatusOK, models.RespInt{
			Code: models.ToCode(models.GetAssetBurnTotalErr),
			Msg:  err.Error(),
			Data: 0,
		})
		return
	}

	c.JSON(http.StatusOK, models.RespInt{
		Code: models.ToCode(models.SUCCESS),
		Msg:  models.NullStr,
		Data: assetBurnTotal.TotalAmount,
	})
	return

}
