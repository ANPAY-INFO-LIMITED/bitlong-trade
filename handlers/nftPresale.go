package handlers

import (
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"trade/btlLog"
	"trade/models"
	"trade/services"
)

// @dev: Get

func GetNftPresaleByAssetId(c *gin.Context) {
	username := c.MustGet("username").(string)
	_, err := services.NameToId(username)
	if err != nil {
		c.JSON(http.StatusOK, models.JsonResult{
			Success: false,
			Error:   err.Error(),
			Code:    models.NameToIdErr,
			Data:    nil,
		})
		return
	}
	assetId := c.Query("asset_id")
	nftPresale, err := services.GetNftPresaleByAssetId(assetId)
	if err != nil {
		c.JSON(http.StatusOK, models.JsonResult{
			Success: false,
			Error:   err.Error(),
			Code:    models.GetNftPresalesByAssetIdErr,
			Data:    nil,
		})
		return
	}
	noMetaStr := c.Query("no_meta")
	noMeta, err := strconv.ParseBool(noMetaStr)
	if err != nil {
		btlLog.PreSale.Error("ParseBool err:%v", err)
	}
	noWhitelistStr := c.Query("no_whitelist")
	noWhitelist, err := strconv.ParseBool(noWhitelistStr)
	if err != nil {
		btlLog.PreSale.Error("ParseBool err:%v", err)
	}
	result := services.NftPresaleToNftPresaleSimplified(nftPresale, noMeta, noWhitelist)
	c.JSON(http.StatusOK, models.JsonResult{
		Success: true,
		Error:   models.SUCCESS.Error(),
		Code:    models.SUCCESS,
		Data:    result,
	})
}

func GetNftPresaleByBatchGroupId(c *gin.Context) {
	username := c.MustGet("username").(string)
	_, err := services.NameToId(username)
	if err != nil {
		c.JSON(http.StatusOK, models.JsonResult{
			Success: false,
			Error:   err.Error(),
			Code:    models.NameToIdErr,
			Data:    nil,
		})
		return
	}
	batchGroupIdStr := c.Query("batch_group_id")
	batchGroupId, err := strconv.Atoi(batchGroupIdStr)
	if err != nil {
		btlLog.PreSale.Error("Atoi err:%v", err)
	}
	nftPresale, err := services.GetNftPresaleByBatchGroupId(batchGroupId)
	if err != nil {
		c.JSON(http.StatusOK, models.JsonResult{
			Success: false,
			Error:   err.Error(),
			Code:    models.GetNftPresaleByBatchGroupIdErr,
			Data:    nil,
		})
		return
	}
	noMetaStr := c.Query("no_meta")
	noMeta, err := strconv.ParseBool(noMetaStr)
	if err != nil {
		btlLog.PreSale.Error("ParseBool err:%v", err)
	}
	noWhitelistStr := c.Query("no_whitelist")
	noWhitelist, err := strconv.ParseBool(noWhitelistStr)
	if err != nil {
		btlLog.PreSale.Error("ParseBool err:%v", err)
	}
	result := services.NftPresaleSliceToNftPresaleSimplifiedSlice(nftPresale, noMeta, noWhitelist)
	c.JSON(http.StatusOK, models.JsonResult{
		Success: true,
		Error:   models.SUCCESS.Error(),
		Code:    models.SUCCESS,
		Data:    result,
	})
}

func GetLaunchedNftPresale(c *gin.Context) {
	username := c.MustGet("username").(string)
	_, err := services.NameToId(username)
	if err != nil {
		c.JSON(http.StatusOK, models.JsonResult{
			Success: false,
			Error:   err.Error(),
			Code:    models.NameToIdErr,
			Data:    nil,
		})
		return
	}
	nftPresales, err := services.GetLaunchedNftPresales()
	if err != nil {
		c.JSON(http.StatusOK, models.JsonResult{
			Success: false,
			Error:   err.Error(),
			Code:    models.GetLaunchedNftPresalesErr,
			Data:    nil,
		})
		return
	}
	noMetaStr := c.Query("no_meta")
	noMeta, err := strconv.ParseBool(noMetaStr)
	if err != nil {
		btlLog.PreSale.Error("ParseBool err:%v", err)
	}
	noWhitelistStr := c.Query("no_whitelist")
	noWhitelist, err := strconv.ParseBool(noWhitelistStr)
	if err != nil {
		btlLog.PreSale.Error("ParseBool err:%v", err)
	}
	result := services.NftPresaleSliceToNftPresaleSimplifiedSlice(nftPresales, noMeta, noWhitelist)
	c.JSON(http.StatusOK, models.JsonResult{
		Success: true,
		Error:   models.SUCCESS.Error(),
		Code:    models.SUCCESS,
		Data:    result,
	})
}

func GetUserBoughtNftPresale(c *gin.Context) {
	username := c.MustGet("username").(string)
	//userId, err := services.NameToId(username)
	//if err != nil {
	//	c.JSON(http.StatusOK, models.JsonResult{
	//		Success: false,
	//		Error:   err.Error(),
	//		Code:    models.NameToIdErr,
	//		Data:    nil,
	//	})
	//	return
	//}
	nftPresales, err := services.GetNftPresalesByBuyerUsername(username)
	if err != nil {
		c.JSON(http.StatusOK, models.JsonResult{
			Success: false,
			Error:   err.Error(),
			Code:    models.GetNftPresalesByBuyerUserIdErr,
			Data:    nil,
		})
		return
	}
	noMetaStr := c.Query("no_meta")
	noMeta, err := strconv.ParseBool(noMetaStr)
	if err != nil {
		btlLog.PreSale.Error("ParseBool err:%v", err)
	}
	noWhitelistStr := c.Query("no_whitelist")
	noWhitelist, err := strconv.ParseBool(noWhitelistStr)
	if err != nil {
		btlLog.PreSale.Error("ParseBool err:%v", err)
	}
	result := services.NftPresaleSliceToNftPresaleSimplifiedSlice(nftPresales, noMeta, noWhitelist)
	c.JSON(http.StatusOK, models.JsonResult{
		Success: true,
		Error:   models.SUCCESS.Error(),
		Code:    models.SUCCESS,
		Data:    result,
	})
}

func GetNftPresaleByGroupKeyPurchasable(c *gin.Context) {
	username := c.MustGet("username").(string)
	_, err := services.NameToId(username)
	if err != nil {
		c.JSON(http.StatusOK, models.JsonResult{
			Success: false,
			Error:   err.Error(),
			Code:    models.NameToIdErr,
			Data:    nil,
		})
		return
	}
	groupKey := c.Query("group_key")
	nftPresales, err := services.GetNftPresaleByGroupKeyPurchasable(groupKey)
	if err != nil {
		c.JSON(http.StatusOK, models.JsonResult{
			Success: false,
			Error:   err.Error(),
			Code:    models.GetNftPresaleByGroupKeyErr,
			Data:    nil,
		})
		return
	}
	noMetaStr := c.Query("no_meta")
	noMeta, err := strconv.ParseBool(noMetaStr)
	if err != nil {
		btlLog.PreSale.Error("ParseBool err:%v", err)
	}
	noWhitelistStr := c.Query("no_whitelist")
	noWhitelist, err := strconv.ParseBool(noWhitelistStr)
	if err != nil {
		btlLog.PreSale.Error("ParseBool err:%v", err)
	}
	result := services.NftPresaleSliceToNftPresaleSimplifiedSlice(nftPresales, noMeta, noWhitelist)
	c.JSON(http.StatusOK, models.JsonResult{
		Success: true,
		Error:   models.SUCCESS.Error(),
		Code:    models.SUCCESS,
		Data:    result,
	})
}

func GetNftPresaleNoGroupKeyPurchasable(c *gin.Context) {
	username := c.MustGet("username").(string)
	_, err := services.NameToId(username)
	if err != nil {
		c.JSON(http.StatusOK, models.JsonResult{
			Success: false,
			Error:   err.Error(),
			Code:    models.NameToIdErr,
			Data:    nil,
		})
		return
	}
	nftPresales, err := services.GetNftPresaleNoGroupKeyPurchasable()
	if err != nil {
		c.JSON(http.StatusOK, models.JsonResult{
			Success: false,
			Error:   err.Error(),
			Code:    models.GetNftPresaleByGroupKeyErr,
			Data:    nil,
		})
		return
	}
	noMetaStr := c.Query("no_meta")
	noMeta, err := strconv.ParseBool(noMetaStr)
	if err != nil {
		btlLog.PreSale.Error("ParseBool err:%v", err)
	}
	noWhitelistStr := c.Query("no_whitelist")
	noWhitelist, err := strconv.ParseBool(noWhitelistStr)
	if err != nil {
		btlLog.PreSale.Error("ParseBool err:%v", err)
	}
	result := services.NftPresaleSliceToNftPresaleSimplifiedSlice(nftPresales, noMeta, noWhitelist)
	c.JSON(http.StatusOK, models.JsonResult{
		Success: true,
		Error:   models.SUCCESS.Error(),
		Code:    models.SUCCESS,
		Data:    result,
	})
}

// @dev: Purchase

func BuyNftPresale(c *gin.Context) {
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
	var buyNftPresaleRequest models.BuyNftPresaleRequest
	err = c.ShouldBindJSON(&buyNftPresaleRequest)
	if err != nil {
		c.JSON(http.StatusOK, models.JsonResult{
			Success: false,
			Error:   err.Error(),
			Code:    models.ShouldBindJsonErr,
			Data:    nil,
		})
		return
	}
	err = services.BuyNftPresale(userId, username, buyNftPresaleRequest)
	if err != nil {
		c.JSON(http.StatusOK, models.JsonResult{
			Success: false,
			Error:   err.Error(),
			Code:    models.BuyNftPresaleErr,
			Data:    nil,
		})
		return
	}
	c.JSON(http.StatusOK, models.JsonResult{
		Success: true,
		Error:   "",
		Code:    models.SUCCESS,
		Data:    nil,
	})
}

// @dev: Query

func QueryNftPresaleGroupKeyPurchasable(c *gin.Context) {
	username := c.MustGet("username").(string)
	_, err := services.NameToId(username)
	if err != nil {
		c.JSON(http.StatusOK, models.JsonResult{
			Success: false,
			Error:   err.Error(),
			Code:    models.NameToIdErr,
			Data:    nil,
		})
		return
	}
	groupKeys, err := services.GetAllNftPresaleGroupKeyPurchasable()
	if err != nil {
		c.JSON(http.StatusOK, models.JsonResult{
			Success: false,
			Error:   err.Error(),
			Code:    models.GetLaunchedNftPresalesErr,
			Data:    nil,
		})
		return
	}
	c.JSON(http.StatusOK, models.JsonResult{
		Success: true,
		Error:   models.SUCCESS.Error(),
		Code:    models.SUCCESS,
		Data:    groupKeys,
	})
}

// @dev: Set

func SetNftPresale(c *gin.Context) {
	var nftPresaleSetRequest models.NftPresaleSetRequest
	err := c.ShouldBindJSON(&nftPresaleSetRequest)
	if err != nil {
		c.JSON(http.StatusOK, models.JsonResult{
			Success: false,
			Error:   err.Error(),
			Code:    models.ShouldBindJsonErr,
			Data:    nil,
		})
		return
	}
	nftPresale := services.ProcessNftPresale(&nftPresaleSetRequest)
	// @dev: Store AssetMeta
	{
		assetId := nftPresaleSetRequest.AssetId
		err = services.StoreAssetMetaIfNotExist(assetId)
		if err != nil {
			btlLog.PreSale.Error("api StoreAssetMetaIfNotExist err:%v", err)
		}
	}
	err = services.CreateNftPresale(nftPresale)
	if err != nil {
		c.JSON(http.StatusOK, models.JsonResult{
			Success: false,
			Error:   err.Error(),
			Code:    models.CreateNftPresaleErr,
			Data:    nil,
		})
		return
	}
	c.JSON(http.StatusOK, models.JsonResult{
		Success: true,
		Error:   "",
		Code:    models.SUCCESS,
		Data:    nil,
	})
}

func SetNftPresales(c *gin.Context) {
	var nftPresaleSetRequests []models.NftPresaleSetRequest
	err := c.ShouldBindJSON(&nftPresaleSetRequests)
	if err != nil {
		c.JSON(http.StatusOK, models.JsonResult{
			Success: false,
			Error:   err.Error(),
			Code:    models.ShouldBindJsonErr,
			Data:    nil,
		})
		return
	}
	nftPresales := services.ProcessNftPresales(&nftPresaleSetRequests)
	// @dev: Store AssetMetas
	{
		var assetIds []string
		for _, nftPresaleSetRequest := range nftPresaleSetRequests {
			assetIds = append(assetIds, nftPresaleSetRequest.AssetId)
		}
		err = services.StoreAssetMetasIfNotExist(assetIds)
		if err != nil {
			btlLog.PreSale.Error("api StoreAssetMetasIfNotExist err:%v", err)
		}
	}
	err = services.CreateNftPresales(nftPresales)
	if err != nil {
		c.JSON(http.StatusOK, models.JsonResult{
			Success: false,
			Error:   err.Error(),
			Code:    models.CreateNftPresalesErr,
			Data:    nil,
		})
		return
	}
	c.JSON(http.StatusOK, models.JsonResult{
		Success: true,
		Error:   "",
		Code:    models.SUCCESS,
		Data:    nil,
	})
}

func ReSetFailOrCanceledNftPresale(c *gin.Context) {
	err := services.ReSetFailOrCanceledNftPresale()
	if err != nil {
		c.JSON(http.StatusOK, models.JsonResult{
			Success: false,
			Error:   err.Error(),
			Code:    models.ReSetFailOrCanceledNftPresaleErr,
			Data:    nil,
		})
		return
	}
	c.JSON(http.StatusOK, models.JsonResult{
		Success: true,
		Error:   "",
		Code:    models.SUCCESS,
		Data:    nil,
	})
}

// @dev: launch batch group

func LaunchNftPresaleBatchGroup(c *gin.Context) {
	var nftPresaleBatchGroupLaunchRequest models.NftPresaleBatchGroupLaunchRequest
	err := c.ShouldBindJSON(&nftPresaleBatchGroupLaunchRequest)
	if err != nil {
		c.JSON(http.StatusOK, models.JsonResult{
			Success: false,
			Error:   err.Error(),
			Code:    models.ShouldBindJsonErr,
			Data:    nil,
		})
		return
	}
	// @dev: Store AssetMetas
	{
		var assetIds []string
		for _, nftPresaleSetRequest := range *(nftPresaleBatchGroupLaunchRequest.NftPresaleSetRequests) {
			assetIds = append(assetIds, nftPresaleSetRequest.AssetId)
		}
		err = services.StoreAssetMetasIfNotExist(assetIds)
		if err != nil {
			btlLog.PreSale.Error("api StoreAssetMetasIfNotExist err:%v", err)
		}
	}
	// @dev: Process and create db records
	err = services.ProcessNftPresaleBatchGroupLaunchRequestAndCreate(&nftPresaleBatchGroupLaunchRequest)
	if err != nil {
		c.JSON(http.StatusOK, models.JsonResult{
			Success: false,
			Error:   err.Error(),
			Code:    models.ProcessNftPresaleBatchGroupLaunchRequestAndCreateErr,
			Data:    nil,
		})
		return
	}
	c.JSON(http.StatusOK, models.JsonResult{
		Success: true,
		Error:   "",
		Code:    models.SUCCESS,
		Data:    nil,
	})
}

func QueryNftPresaleBatchGroup(c *gin.Context) {
	username := c.MustGet("username").(string)
	_, err := services.NameToId(username)
	if err != nil {
		c.JSON(http.StatusOK, models.JsonResult{
			Success: false,
			Error:   err.Error(),
			Code:    models.NameToIdErr,
			Data:    nil,
		})
		return
	}
	stateStr := c.Query("state")
	state, err := strconv.Atoi(stateStr)
	if err != nil {
		btlLog.PreSale.Error("Atoi err:%v", err)
	}
	batchGroups, err := services.GetBatchGroups(models.NftPresaleBatchGroupState(state))
	if err != nil {
		c.JSON(http.StatusOK, models.JsonResult{
			Success: false,
			Error:   err.Error(),
			Code:    models.GetBatchGroupsErr,
			Data:    nil,
		})
		return
	}
	result := services.NftPresaleBatchGroupSliceToNftPresaleBatchGroupSimplifiedSlice(batchGroups)
	c.JSON(http.StatusOK, models.JsonResult{
		Success: true,
		Error:   models.SUCCESS.Error(),
		Code:    models.SUCCESS,
		Data:    result,
	})
}

// router second

func GetPurchasedNftPresaleInfo(c *gin.Context) {
	nftPresaleInfos, err := services.GetPurchasedNftPresaleInfo()
	if err != nil {
		c.JSON(http.StatusOK, Result2{
			Errno:  models.GetPurchasedNftPresaleInfoErr.Code(),
			ErrMsg: err.Error(),
			Data:   nftPresaleInfos,
		})
		return
	}
	c.JSON(http.StatusOK, Result2{
		Errno:  0,
		ErrMsg: models.SUCCESS.Error(),
		Data:   nftPresaleInfos,
	})
}

func GetPurchasedNftPresaleInfoPageAndRows(c *gin.Context) {
	page := c.Query("page")
	rows := c.Query("rows")

	var err error
	var pageInt, rowsInt int
	var count int64

	if page == "" {
		err = errors.New("page is empty")
		c.JSON(http.StatusOK, Result2{
			Errno:  models.PageEmptyErr.Code(),
			ErrMsg: err.Error(),
			Data: ResultData{
				Page:      pageInt,
				Rows:      rowsInt,
				Count:     count,
				DataSlice: []services.NftPresaleInfo{},
			},
		})
		return
	}
	pageInt, err = strconv.Atoi(page)
	if err != nil {
		c.JSON(http.StatusOK, Result2{
			Errno:  models.AtoiErr.Code(),
			ErrMsg: err.Error(),
			Data: ResultData{
				Page:      pageInt,
				Rows:      rowsInt,
				Count:     count,
				DataSlice: []services.NftPresaleInfo{},
			},
		})
		return
	}
	if pageInt < 0 {
		err = errors.New("page is less than 0")
		c.JSON(http.StatusOK, Result2{
			Errno:  models.PageLessThanZeroErr.Code(),
			ErrMsg: err.Error(),
			Data: ResultData{
				Page:      pageInt,
				Rows:      rowsInt,
				Count:     count,
				DataSlice: []services.NftPresaleInfo{},
			},
		})
		return
	}

	if rows == "" {
		err := errors.New("rows is empty")
		c.JSON(http.StatusOK, Result2{
			Errno:  models.RowsEmptyErr.Code(),
			ErrMsg: err.Error(),
			Data: ResultData{
				Page:      pageInt,
				Rows:      rowsInt,
				Count:     count,
				DataSlice: []services.NftPresaleInfo{},
			},
		})
		return
	}
	rowsInt, err = strconv.Atoi(rows)
	if err != nil {
		c.JSON(http.StatusOK, Result2{
			Errno:  models.AtoiErr.Code(),
			ErrMsg: err.Error(),
			Data: ResultData{
				Page:      pageInt,
				Rows:      rowsInt,
				Count:     count,
				DataSlice: []services.NftPresaleInfo{},
			},
		})
		return
	}
	if rowsInt < 0 {
		err = errors.New("rows is less than 0")
		c.JSON(http.StatusOK, Result2{
			Errno:  models.RowsLessThanZeroErr.Code(),
			ErrMsg: err.Error(),
			Data: ResultData{
				Page:      pageInt,
				Rows:      rowsInt,
				Count:     count,
				DataSlice: []services.NftPresaleInfo{},
			},
		})
		return
	}

	limit := rowsInt
	offset := (pageInt - 1) * rowsInt

	if offset < 0 {
		err = errors.New("offset is less than 0")
		c.JSON(http.StatusOK, Result2{
			Errno:  models.OffsetLessThanZeroErr.Code(),
			ErrMsg: err.Error(),
			Data: ResultData{
				Page:      pageInt,
				Rows:      rowsInt,
				Count:     count,
				DataSlice: []services.NftPresaleInfo{},
			},
		})
		return
	}

	count, err = services.GetPurchasedNftPresaleInfoCount()
	if err != nil {
		c.JSON(http.StatusOK, Result2{
			Errno:  models.GetPurchasedNftPresaleInfoCountErr.Code(),
			ErrMsg: err.Error(),
			Data: ResultData{
				Page:      pageInt,
				Rows:      rowsInt,
				Count:     count,
				DataSlice: []services.NftPresaleInfo{},
			},
		})
		return
	}

	nftPresaleInfos, err := services.GetPurchasedNftPresaleInfoLimitAndOffset(limit, offset)
	if err != nil {
		c.JSON(http.StatusOK, Result2{
			Errno:  models.GetPurchasedNftPresaleInfoLimitAndOffsetErr.Code(),
			ErrMsg: err.Error(),
			Data: ResultData{
				Page:      pageInt,
				Rows:      rowsInt,
				Count:     count,
				DataSlice: nftPresaleInfos,
			},
		})
		return
	}
	c.JSON(http.StatusOK, Result2{
		Errno:  0,
		ErrMsg: models.SUCCESS.Error(),
		Data: ResultData{
			Page:      pageInt,
			Rows:      rowsInt,
			Count:     count,
			DataSlice: nftPresaleInfos,
		},
	})
}

// Offline purchase data

func GetNftPresaleOfflinePurchaseData(c *gin.Context) {
	nftNo := c.Query("nft_no")
	npubKey := c.Query("npub_key")
	invitationCode := c.Query("invitation_code")
	assetId := c.Query("asset_id")
	nftPresaleOfflinePurchaseDataInfos, err := services.GetNftPresaleOfflinePurchaseData(nftNo, npubKey, invitationCode, assetId)
	if err != nil {
		c.JSON(http.StatusOK, Result2{
			Errno:  models.GetNftPresaleOfflinePurchaseDataErr.Code(),
			ErrMsg: err.Error(),
			Data:   new([]services.NftPresaleOfflinePurchaseDataInfo),
		})
		return
	}
	c.JSON(http.StatusOK, Result2{
		Errno:  0,
		ErrMsg: models.SUCCESS.Error(),
		Data:   nftPresaleOfflinePurchaseDataInfos,
	})
}

func UpdateNftPresaleOfflinePurchaseData(c *gin.Context) {
	var request services.UpdateNftPresaleOfflinePurchaseDataRequest
	err := c.ShouldBindJSON(&request)
	if err != nil {
		c.JSON(http.StatusOK, Result2{
			Errno:  models.ShouldBindJsonErr.Code(),
			ErrMsg: err.Error(),
			Data:   nil,
		})
		return
	}
	err = services.UpdateNftPresaleOfflinePurchaseData(request.NftNo, request.NpubKey, request.InvitationCode, request.Name)
	if err != nil {
		c.JSON(http.StatusOK, Result2{
			Errno:  models.UpdateNftPresaleOfflinePurchaseDataErr.Code(),
			ErrMsg: err.Error(),
			Data:   nil,
		})
		return
	}
	c.JSON(http.StatusOK, Result2{
		Errno:  0,
		ErrMsg: models.SUCCESS.Error(),
		Data:   nil,
	})
}
