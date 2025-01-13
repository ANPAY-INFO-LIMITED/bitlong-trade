package handlers

import (
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"trade/models"
	"trade/services/assetBalanceBackend"
)

func GetAssetBalanceLimitAndOffset(c *gin.Context) {
	assetId := c.Query("asset_id")
	limit := c.Query("limit")
	offset := c.Query("offset")

	if assetId == "" {
		err := errors.New("asset_id is empty")
		c.JSON(http.StatusOK, Result2{
			Errno:  models.AssetIdEmptyErr.Code(),
			ErrMsg: err.Error(),
			Data:   &[]assetBalanceBackend.AssetBalanceInfo{},
		})
		return
	}

	if limit == "" {
		err := errors.New("limit is empty")
		c.JSON(http.StatusOK, Result2{
			Errno:  models.LimitEmptyErr.Code(),
			ErrMsg: err.Error(),
			Data:   &[]assetBalanceBackend.AssetBalanceInfo{},
		})
		return
	}
	limitInt, err := strconv.Atoi(limit)
	if err != nil {
		c.JSON(http.StatusOK, Result2{
			Errno:  models.AtoiErr.Code(),
			ErrMsg: err.Error(),
			Data:   &[]assetBalanceBackend.AssetBalanceInfo{},
		})
		return
	}
	if limitInt < 0 {
		err := errors.New("limit is less than 0")
		c.JSON(http.StatusOK, Result2{
			Errno:  models.LimitLessThanZeroErr.Code(),
			ErrMsg: err.Error(),
			Data:   &[]assetBalanceBackend.AssetBalanceInfo{},
		})
		return
	}

	if offset == "" {
		err := errors.New("offset is empty")
		c.JSON(http.StatusOK, Result2{
			Errno:  models.OffsetEmptyErr.Code(),
			ErrMsg: err.Error(),
			Data:   &[]assetBalanceBackend.AssetBalanceInfo{},
		})
		return
	}
	offsetInt, err := strconv.Atoi(offset)
	if err != nil {
		c.JSON(http.StatusOK, Result2{
			Errno:  models.AtoiErr.Code(),
			ErrMsg: err.Error(),
			Data:   &[]assetBalanceBackend.AssetBalanceInfo{},
		})
		return
	}
	if offsetInt < 0 {
		err := errors.New("offset is less than 0")
		c.JSON(http.StatusOK, Result2{
			Errno:  models.OffsetLessThanZeroErr.Code(),
			ErrMsg: err.Error(),
			Data:   &[]assetBalanceBackend.AssetBalanceInfo{},
		})
		return
	}

	assetBalanceInfos, err := assetBalanceBackend.GetAssetBalanceLimitAndOffset(assetId, limitInt, offsetInt)

	if err != nil {
		c.JSON(http.StatusOK, Result2{
			Errno:  models.GetAssetBalanceLimitAndOffsetErr.Code(),
			ErrMsg: err.Error(),
			Data:   &[]assetBalanceBackend.AssetBalanceInfo{},
		})
		return
	}

	c.JSON(http.StatusOK, Result2{
		Errno:  0,
		ErrMsg: models.SUCCESS.Error(),
		Data:   assetBalanceInfos,
	})
}

func GetAssetBalanceCount(c *gin.Context) {
	assetId := c.Query("asset_id")
	if assetId == "" {
		err := errors.New("asset_id is empty")
		c.JSON(http.StatusOK, Result2{
			Errno:  models.AssetIdEmptyErr.Code(),
			ErrMsg: err.Error(),
			Data:   0,
		})
		return
	}
	count, err := assetBalanceBackend.GetAssetBalanceCount(assetId)

	if err != nil {
		c.JSON(http.StatusOK, Result2{
			Errno:  models.GetAssetBalanceCountErr.Code(),
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

func QueryAssetBalanceInfoByUsername(c *gin.Context) {
	assetId := c.Query("asset_id")
	if assetId == "" {
		err := errors.New("asset_id is empty")
		c.JSON(http.StatusOK, Result2{
			Errno:  models.AssetIdEmptyErr.Code(),
			ErrMsg: err.Error(),
			Data:   0,
		})
		return
	}
	username := c.Query("username")
	if username == "" {
		err := errors.New("username is empty")
		c.JSON(http.StatusOK, Result2{
			Errno:  models.UsernameEmptyErr.Code(),
			ErrMsg: err.Error(),
			Data:   0,
		})
		return
	}

	count, err := assetBalanceBackend.QueryAssetBalanceInfoByUsername(assetId, username)

	if err != nil {
		c.JSON(http.StatusOK, Result2{
			Errno:  models.QueryAssetBalanceInfoByUsernameErr.Code(),
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

func QueryAllAssetBalanceAssetIds(c *gin.Context) {
	assetIds, err := assetBalanceBackend.QueryAllAssetBalanceAssetIds()

	if err != nil {
		c.JSON(http.StatusOK, Result2{
			Errno:  models.GetAssetBalanceCountErr.Code(),
			ErrMsg: err.Error(),
			Data:   []string{},
		})
		return
	}

	c.JSON(http.StatusOK, Result2{
		Errno:  0,
		ErrMsg: models.SUCCESS.Error(),
		Data:   assetIds,
	})
}

func GetAssetBalanceHistoryLimitAndOffset(c *gin.Context) {
	assetId := c.Query("asset_id")
	limit := c.Query("limit")
	offset := c.Query("offset")

	if assetId == "" {
		err := errors.New("asset_id is empty")
		c.JSON(http.StatusOK, Result2{
			Errno:  models.AssetIdEmptyErr.Code(),
			ErrMsg: err.Error(),
			Data:   &[]assetBalanceBackend.AssetBalanceHistoryInfo{},
		})
		return
	}

	if limit == "" {
		err := errors.New("limit is empty")
		c.JSON(http.StatusOK, Result2{
			Errno:  models.LimitEmptyErr.Code(),
			ErrMsg: err.Error(),
			Data:   &[]assetBalanceBackend.AssetBalanceHistoryInfo{},
		})
		return
	}
	limitInt, err := strconv.Atoi(limit)
	if err != nil {
		c.JSON(http.StatusOK, Result2{
			Errno:  models.AtoiErr.Code(),
			ErrMsg: err.Error(),
			Data:   &[]assetBalanceBackend.AssetBalanceHistoryInfo{},
		})
		return
	}
	if limitInt < 0 {
		err := errors.New("limit is less than 0")
		c.JSON(http.StatusOK, Result2{
			Errno:  models.LimitLessThanZeroErr.Code(),
			ErrMsg: err.Error(),
			Data:   &[]assetBalanceBackend.AssetBalanceHistoryInfo{},
		})
		return
	}

	if offset == "" {
		err := errors.New("offset is empty")
		c.JSON(http.StatusOK, Result2{
			Errno:  models.OffsetEmptyErr.Code(),
			ErrMsg: err.Error(),
			Data:   &[]assetBalanceBackend.AssetBalanceHistoryInfo{},
		})
		return
	}
	offsetInt, err := strconv.Atoi(offset)
	if err != nil {
		c.JSON(http.StatusOK, Result2{
			Errno:  models.AtoiErr.Code(),
			ErrMsg: err.Error(),
			Data:   &[]assetBalanceBackend.AssetBalanceHistoryInfo{},
		})
		return
	}
	if offsetInt < 0 {
		err := errors.New("offset is less than 0")
		c.JSON(http.StatusOK, Result2{
			Errno:  models.OffsetLessThanZeroErr.Code(),
			ErrMsg: err.Error(),
			Data:   &[]assetBalanceBackend.AssetBalanceHistoryInfo{},
		})
		return
	}

	AssetBalanceHistoryInfos, err := assetBalanceBackend.GetAssetBalanceHistoryLimitAndOffset(assetId, limitInt, offsetInt)

	if err != nil {
		c.JSON(http.StatusOK, Result2{
			Errno:  models.GetAssetBalanceLimitAndOffsetErr.Code(),
			ErrMsg: err.Error(),
			Data:   &[]assetBalanceBackend.AssetBalanceHistoryInfo{},
		})
		return
	}

	c.JSON(http.StatusOK, Result2{
		Errno:  0,
		ErrMsg: models.SUCCESS.Error(),
		Data:   AssetBalanceHistoryInfos,
	})
}

func GetAssetBalanceHistoryCount(c *gin.Context) {
	assetId := c.Query("asset_id")
	if assetId == "" {
		err := errors.New("asset_id is empty")
		c.JSON(http.StatusOK, Result2{
			Errno:  models.AssetIdEmptyErr.Code(),
			ErrMsg: err.Error(),
			Data:   0,
		})
		return
	}
	count, err := assetBalanceBackend.GetAssetBalanceHistoryCount(assetId)

	if err != nil {
		c.JSON(http.StatusOK, Result2{
			Errno:  models.GetAssetBalanceHistoryCountErr.Code(),
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

func QueryAssetBalanceHistoryInfoByUsername(c *gin.Context) {
	assetId := c.Query("asset_id")
	if assetId == "" {
		err := errors.New("asset_id is empty")
		c.JSON(http.StatusOK, Result2{
			Errno:  models.AssetIdEmptyErr.Code(),
			ErrMsg: err.Error(),
			Data:   0,
		})
		return
	}
	username := c.Query("username")
	if username == "" {
		err := errors.New("username is empty")
		c.JSON(http.StatusOK, Result2{
			Errno:  models.UsernameEmptyErr.Code(),
			ErrMsg: err.Error(),
			Data:   0,
		})
		return
	}

	count, err := assetBalanceBackend.QueryAssetBalanceHistoryInfoByUsername(assetId, username)

	if err != nil {
		c.JSON(http.StatusOK, Result2{
			Errno:  models.QueryAssetBalanceHistoryInfoByUsernameErr.Code(),
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

func QueryAllAssetBalanceHistoryAssetIds(c *gin.Context) {
	assetIds, err := assetBalanceBackend.QueryAllAssetBalanceHistoryAssetIds()

	if err != nil {
		c.JSON(http.StatusOK, Result2{
			Errno:  models.GetAssetBalanceCountErr.Code(),
			ErrMsg: err.Error(),
			Data:   []string{},
		})
		return
	}

	c.JSON(http.StatusOK, Result2{
		Errno:  0,
		ErrMsg: models.SUCCESS.Error(),
		Data:   assetIds,
	})
}
