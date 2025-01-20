package handlers

import (
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"trade/models"
	"trade/services/assetMoreInfo"
)

//链上持有

func GetAssetBalanceInfoCount(c *gin.Context) {
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

	count, err := assetMoreInfo.GetAssetBalanceInfoCount(assetId)
	if err != nil {
		c.JSON(http.StatusOK, Result2{
			Errno:  models.GetAssetBalanceInfoCountErr.Code(),
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

func GetAssetBalanceInfo(c *gin.Context) {
	assetId := c.Query("asset_id")
	limit := c.Query("limit")
	offset := c.Query("offset")

	if assetId == "" {
		err := errors.New("asset_id is empty")
		c.JSON(http.StatusOK, Result2{
			Errno:  models.AssetIdEmptyErr.Code(),
			ErrMsg: err.Error(),
			Data:   []assetMoreInfo.AssetBalanceInfo{},
		})
		return
	}

	if limit == "" {
		err := errors.New("limit is empty")
		c.JSON(http.StatusOK, Result2{
			Errno:  models.LimitEmptyErr.Code(),
			ErrMsg: err.Error(),
			Data:   []assetMoreInfo.AssetBalanceInfo{},
		})
		return
	}
	limitInt, err := strconv.Atoi(limit)
	if err != nil {
		c.JSON(http.StatusOK, Result2{
			Errno:  models.AtoiErr.Code(),
			ErrMsg: err.Error(),
			Data:   []assetMoreInfo.AssetBalanceInfo{},
		})
		return
	}
	if limitInt < 0 {
		err := errors.New("limit is less than 0")
		c.JSON(http.StatusOK, Result2{
			Errno:  models.LimitLessThanZeroErr.Code(),
			ErrMsg: err.Error(),
			Data:   []assetMoreInfo.AssetBalanceInfo{},
		})
		return
	}

	if offset == "" {
		err := errors.New("offset is empty")
		c.JSON(http.StatusOK, Result2{
			Errno:  models.OffsetEmptyErr.Code(),
			ErrMsg: err.Error(),
			Data:   []assetMoreInfo.AssetBalanceInfo{},
		})
		return
	}
	offsetInt, err := strconv.Atoi(offset)
	if err != nil {
		c.JSON(http.StatusOK, Result2{
			Errno:  models.AtoiErr.Code(),
			ErrMsg: err.Error(),
			Data:   []assetMoreInfo.AssetBalanceInfo{},
		})
		return
	}
	if offsetInt < 0 {
		err := errors.New("offset is less than 0")
		c.JSON(http.StatusOK, Result2{
			Errno:  models.OffsetLessThanZeroErr.Code(),
			ErrMsg: err.Error(),
			Data:   []assetMoreInfo.AssetBalanceInfo{},
		})
		return
	}

	assetBalanceInfos, err := assetMoreInfo.GetAssetBalanceInfo(assetId, limitInt, offsetInt)

	if err != nil {
		c.JSON(http.StatusOK, Result2{
			Errno:  models.GetAssetBalanceInfoErr.Code(),
			ErrMsg: err.Error(),
			Data:   &[]assetMoreInfo.AssetBalanceInfo{},
		})
		return
	}

	c.JSON(http.StatusOK, Result2{
		Errno:  0,
		ErrMsg: models.SUCCESS.Error(),
		Data:   assetBalanceInfos,
	})
}

//通道记录

func GetAccountAssetTransferCount(c *gin.Context) {
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

	count, err := assetMoreInfo.GetAccountAssetTransferCount(assetId)
	if err != nil {
		c.JSON(http.StatusOK, Result2{
			Errno:  models.GetAccountAssetTransferCountErr.Code(),
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

func GetAccountAssetTransfer(c *gin.Context) {
	assetId := c.Query("asset_id")
	limit := c.Query("limit")
	offset := c.Query("offset")

	if assetId == "" {
		err := errors.New("asset_id is empty")
		c.JSON(http.StatusOK, Result2{
			Errno:  models.AssetIdEmptyErr.Code(),
			ErrMsg: err.Error(),
			Data:   []assetMoreInfo.AccountAssetTransfer{},
		})
		return
	}

	if limit == "" {
		err := errors.New("limit is empty")
		c.JSON(http.StatusOK, Result2{
			Errno:  models.LimitEmptyErr.Code(),
			ErrMsg: err.Error(),
			Data:   []assetMoreInfo.AccountAssetTransfer{},
		})
		return
	}
	limitInt, err := strconv.Atoi(limit)
	if err != nil {
		c.JSON(http.StatusOK, Result2{
			Errno:  models.AtoiErr.Code(),
			ErrMsg: err.Error(),
			Data:   []assetMoreInfo.AccountAssetTransfer{},
		})
		return
	}
	if limitInt < 0 {
		err := errors.New("limit is less than 0")
		c.JSON(http.StatusOK, Result2{
			Errno:  models.LimitLessThanZeroErr.Code(),
			ErrMsg: err.Error(),
			Data:   []assetMoreInfo.AccountAssetTransfer{},
		})
		return
	}

	if offset == "" {
		err := errors.New("offset is empty")
		c.JSON(http.StatusOK, Result2{
			Errno:  models.OffsetEmptyErr.Code(),
			ErrMsg: err.Error(),
			Data:   []assetMoreInfo.AccountAssetTransfer{},
		})
		return
	}
	offsetInt, err := strconv.Atoi(offset)
	if err != nil {
		c.JSON(http.StatusOK, Result2{
			Errno:  models.AtoiErr.Code(),
			ErrMsg: err.Error(),
			Data:   []assetMoreInfo.AccountAssetTransfer{},
		})
		return
	}
	if offsetInt < 0 {
		err := errors.New("offset is less than 0")
		c.JSON(http.StatusOK, Result2{
			Errno:  models.OffsetLessThanZeroErr.Code(),
			ErrMsg: err.Error(),
			Data:   []assetMoreInfo.AccountAssetTransfer{},
		})
		return
	}

	accountAssetTransfers, err := assetMoreInfo.GetAccountAssetTransfer(assetId, limitInt, offsetInt)

	if err != nil {
		c.JSON(http.StatusOK, Result2{
			Errno:  models.GetAccountAssetTransferErr.Code(),
			ErrMsg: err.Error(),
			Data:   &[]assetMoreInfo.AccountAssetTransfer{},
		})
		return
	}

	c.JSON(http.StatusOK, Result2{
		Errno:  0,
		ErrMsg: models.SUCCESS.Error(),
		Data:   accountAssetTransfers,
	})
}

//UTXO

func GetAssetManagedUtxoInfoCount(c *gin.Context) {
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

	count, err := assetMoreInfo.GetAssetBalanceInfoCount(assetId)
	if err != nil {
		c.JSON(http.StatusOK, Result2{
			Errno:  models.GetAssetManagedUtxoInfoCountErr.Code(),
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

func GetAssetManagedUtxoInfo(c *gin.Context) {
	assetId := c.Query("asset_id")
	limit := c.Query("limit")
	offset := c.Query("offset")

	if assetId == "" {
		err := errors.New("asset_id is empty")
		c.JSON(http.StatusOK, Result2{
			Errno:  models.AssetIdEmptyErr.Code(),
			ErrMsg: err.Error(),
			Data:   []assetMoreInfo.AssetManagedUtxoInfo{},
		})
		return
	}

	if limit == "" {
		err := errors.New("limit is empty")
		c.JSON(http.StatusOK, Result2{
			Errno:  models.LimitEmptyErr.Code(),
			ErrMsg: err.Error(),
			Data:   []assetMoreInfo.AssetManagedUtxoInfo{},
		})
		return
	}
	limitInt, err := strconv.Atoi(limit)
	if err != nil {
		c.JSON(http.StatusOK, Result2{
			Errno:  models.AtoiErr.Code(),
			ErrMsg: err.Error(),
			Data:   []assetMoreInfo.AssetManagedUtxoInfo{},
		})
		return
	}
	if limitInt < 0 {
		err := errors.New("limit is less than 0")
		c.JSON(http.StatusOK, Result2{
			Errno:  models.LimitLessThanZeroErr.Code(),
			ErrMsg: err.Error(),
			Data:   []assetMoreInfo.AssetManagedUtxoInfo{},
		})
		return
	}

	if offset == "" {
		err := errors.New("offset is empty")
		c.JSON(http.StatusOK, Result2{
			Errno:  models.OffsetEmptyErr.Code(),
			ErrMsg: err.Error(),
			Data:   []assetMoreInfo.AssetManagedUtxoInfo{},
		})
		return
	}
	offsetInt, err := strconv.Atoi(offset)
	if err != nil {
		c.JSON(http.StatusOK, Result2{
			Errno:  models.AtoiErr.Code(),
			ErrMsg: err.Error(),
			Data:   []assetMoreInfo.AssetManagedUtxoInfo{},
		})
		return
	}
	if offsetInt < 0 {
		err := errors.New("offset is less than 0")
		c.JSON(http.StatusOK, Result2{
			Errno:  models.OffsetLessThanZeroErr.Code(),
			ErrMsg: err.Error(),
			Data:   []assetMoreInfo.AssetManagedUtxoInfo{},
		})
		return
	}

	assetManagedUtxoInfos, err := assetMoreInfo.GetAssetManagedUtxoInfo(assetId, limitInt, offsetInt)

	if err != nil {
		c.JSON(http.StatusOK, Result2{
			Errno:  models.GetAssetManagedUtxoInfoErr.Code(),
			ErrMsg: err.Error(),
			Data:   &[]assetMoreInfo.AssetManagedUtxoInfo{},
		})
		return
	}

	c.JSON(http.StatusOK, Result2{
		Errno:  0,
		ErrMsg: models.SUCCESS.Error(),
		Data:   assetManagedUtxoInfos,
	})
}

//交易

func GetAssetTransferCombinedSliceByAssetIdLimit(c *gin.Context) {
	assetId := c.Query("asset_id")
	if assetId == "" {
		err := errors.New("asset_id is empty")
		c.JSON(http.StatusOK, Result2{
			Errno:  models.AssetIdEmptyErr.Code(),
			ErrMsg: err.Error(),
			Data:   &[]models.AssetTransferProcessedCombined{},
		})
		return
	}

	assetTransfers, err := assetMoreInfo.GetAssetTransferCombinedSliceByAssetIdLimit(assetId, 50)

	if err != nil {
		c.JSON(http.StatusOK, Result2{
			Errno:  models.GetAssetTransferCombinedSliceByAssetIdLimitErr.Code(),
			ErrMsg: err.Error(),
			Data:   &[]models.AssetTransferProcessedCombined{},
		})
		return
	}

	c.JSON(http.StatusOK, Result2{
		Errno:  0,
		ErrMsg: models.SUCCESS.Error(),
		Data:   assetTransfers,
	})
}

//销毁

func GetAssetBurnTotal(c *gin.Context) {
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

	assetBurnTotalAmount, err := assetMoreInfo.GetAssetBurnTotal(assetId)

	if err != nil {
		c.JSON(http.StatusOK, Result2{
			Errno:  models.GetAssetBurnTotalErr.Code(),
			ErrMsg: err.Error(),
			Data:   0,
		})
		return
	}

	c.JSON(http.StatusOK, Result2{
		Errno:  0,
		ErrMsg: models.SUCCESS.Error(),
		Data:   assetBurnTotalAmount,
	})
}
