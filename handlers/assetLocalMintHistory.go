package handlers

import (
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"trade/models"
	"trade/services"
)

func GetAssetLocalMintHistoryByUserId(c *gin.Context) {
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
	assetLocalMintHistories, err := services.GetAssetLocalMintHistoriesByUserId(userId)
	if err != nil {
		c.JSON(http.StatusOK, models.JsonResult{
			Success: false,
			Error:   err.Error(),
			Code:    models.GetAssetLocalMintHistoriesByUserIdErr,
			Data:    nil,
		})
		return
	}
	c.JSON(http.StatusOK, models.JsonResult{
		Success: true,
		Error:   models.SuccessErr,
		Code:    models.SUCCESS,
		Data:    assetLocalMintHistories,
	})
}

func GetAssetLocalMintHistoryAssetId(c *gin.Context) {
	assetId := c.Param("asset_id")
	assetLocalMintHistory, err := services.GetAssetLocalMintHistoryByAssetId(assetId)
	if err != nil {
		c.JSON(http.StatusOK, models.JsonResult{
			Success: false,
			Error:   err.Error(),
			Code:    models.GetAssetLocalMintHistoryByAssetIdErr,
			Data:    nil,
		})
		return
	}
	c.JSON(http.StatusOK, models.JsonResult{
		Success: true,
		Error:   models.SuccessErr,
		Code:    models.SUCCESS,
		Data:    assetLocalMintHistory,
	})
}

func SetAssetLocalMintHistory(c *gin.Context) {
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
	var assetLocalMintHistorySetRequest models.AssetLocalMintHistorySetRequest
	err = c.ShouldBindJSON(&assetLocalMintHistorySetRequest)
	if err != nil {
		c.JSON(http.StatusOK, models.JsonResult{
			Success: false,
			Error:   err.Error(),
			Code:    models.ShouldBindJsonErr,
			Data:    nil,
		})
		return
	}
	assetLocalMintHistory := services.ProcessAssetLocalMintHistorySetRequest(userId, username, assetLocalMintHistorySetRequest)
	err = services.CreateOrUpdateAssetLocalMintHistory(userId, &assetLocalMintHistory)
	if err != nil {
		c.JSON(http.StatusOK, models.JsonResult{
			Success: false,
			Error:   err.Error(),
			Code:    models.SetAssetLocalMintHistoryErr,
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

func SetAssetLocalMintHistories(c *gin.Context) {
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
	var assetLocalMintHistorySetRequests []models.AssetLocalMintHistorySetRequest
	err = c.ShouldBindJSON(&assetLocalMintHistorySetRequests)
	if err != nil {
		c.JSON(http.StatusOK, models.JsonResult{
			Success: false,
			Error:   err.Error(),
			Code:    models.ShouldBindJsonErr,
			Data:    nil,
		})
		return
	}
	assetLocalMintHistories := services.ProcessAssetLocalMintHistorySetRequests(userId, username, &assetLocalMintHistorySetRequests)
	err = services.CreateOrUpdateAssetLocalMintHistories(userId, assetLocalMintHistories)
	if err != nil {
		c.JSON(http.StatusOK, models.JsonResult{
			Success: false,
			Error:   err.Error(),
			Code:    models.SetAssetLocalMintHistoriesErr,
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

func GetAllAssetLocalMintHistorySimplified(c *gin.Context) {
	assetLocalMintHistorySimplified, err := services.GetAllAssetLocalMintHistorySimplified()
	if err != nil {
		c.JSON(http.StatusOK, models.JsonResult{
			Success: false,
			Error:   err.Error(),
			Code:    models.GetAllAssetLocalMintHistorySimplifiedErr,
			Data:    nil,
		})
		return
	}
	c.JSON(http.StatusOK, models.JsonResult{
		Success: true,
		Error:   models.SuccessErr,
		Code:    models.SUCCESS,
		Data:    assetLocalMintHistorySimplified,
	})
}

func GetAssetLocalMintHistoryInfoCount(c *gin.Context) {
	count, err := services.GetAssetLocalMintHistoryInfoCount()
	if err != nil {
		c.JSON(http.StatusOK, Result2{
			Errno:  models.GetAssetLocalMintHistoryInfoCountErr.Code(),
			ErrMsg: err.Error(),
			Data:   0,
		})
		return
	}

	c.JSON(http.StatusOK, Result2{
		Errno:  0,
		ErrMsg: models.SUCCESS.Error(),
		Data:   count,
	})
}

func GetAssetLocalMintHistoryInfo(c *gin.Context) {
	limit := c.Query("limit")
	offset := c.Query("offset")

	if limit == "" {
		err := errors.New("limit is empty")
		c.JSON(http.StatusOK, Result2{
			Errno:  models.LimitEmptyErr.Code(),
			ErrMsg: err.Error(),
			Data:   []services.AssetLocalMintInfo{},
		})
		return
	}
	limitInt, err := strconv.Atoi(limit)
	if err != nil {
		c.JSON(http.StatusOK, Result2{
			Errno:  models.AtoiErr.Code(),
			ErrMsg: err.Error(),
			Data:   []services.AssetLocalMintInfo{},
		})
		return
	}
	if limitInt < 0 {
		err := errors.New("limit is less than 0")
		c.JSON(http.StatusOK, Result2{
			Errno:  models.LimitLessThanZeroErr.Code(),
			ErrMsg: err.Error(),
			Data:   []services.AssetLocalMintInfo{},
		})
		return
	}

	if offset == "" {
		err := errors.New("offset is empty")
		c.JSON(http.StatusOK, Result2{
			Errno:  models.OffsetEmptyErr.Code(),
			ErrMsg: err.Error(),
			Data:   []services.AssetLocalMintInfo{},
		})
		return
	}
	offsetInt, err := strconv.Atoi(offset)
	if err != nil {
		c.JSON(http.StatusOK, Result2{
			Errno:  models.AtoiErr.Code(),
			ErrMsg: err.Error(),
			Data:   []services.AssetLocalMintInfo{},
		})
		return
	}
	if offsetInt < 0 {
		err := errors.New("offset is less than 0")
		c.JSON(http.StatusOK, Result2{
			Errno:  models.OffsetLessThanZeroErr.Code(),
			ErrMsg: err.Error(),
			Data:   []services.AssetLocalMintInfo{},
		})
		return
	}

	assetLocalMintInfos, err := services.GetAssetLocalMintHistoryInfo(limitInt, offsetInt)

	if err != nil {
		c.JSON(http.StatusOK, Result2{
			Errno:  models.GetAssetLocalMintHistoryInfoErr.Code(),
			ErrMsg: err.Error(),
			Data:   &[]services.AssetLocalMintInfo{},
		})
		return
	}

	c.JSON(http.StatusOK, Result2{
		Errno:  0,
		ErrMsg: models.SUCCESS.Error(),
		Data:   assetLocalMintInfos,
	})
}
