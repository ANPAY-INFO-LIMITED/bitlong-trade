package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"net/http"
	"strconv"
	"trade/btlLog"
	"trade/models"
	"trade/services"
	"trade/services/backmanage/ManageQuery"
)

func GetAccountAssetBalanceByAssetId(c *gin.Context) {
	assetId := c.Param("asset_id")
	accountAssetBalanceExtends, err := services.GetAccountAssetBalanceExtendsByAssetId(assetId)
	if err != nil {
		c.JSON(http.StatusOK, models.JsonResult{
			Success: false,
			Error:   err.Error(),
			Code:    models.GetAccountAssetBalanceExtendsByAssetIdErr,
			Data:    nil,
		})
		return
	}
	c.JSON(http.StatusOK, models.JsonResult{
		Success: true,
		Error:   models.SUCCESS.Error(),
		Code:    models.SUCCESS,
		Data:    accountAssetBalanceExtends,
	})
}

func GetAllAccountAssetTransferByAssetId(c *gin.Context) {
	assetId := c.Param("asset_id")
	accountAssetTransfers, err := services.GetAllAccountAssetTransfersByAssetId(assetId)
	if err != nil {
		c.JSON(http.StatusOK, models.JsonResult{
			Success: false,
			Error:   err.Error(),
			Code:    models.GetAllAccountAssetTransfersByAssetIdErr,
			Data:    nil,
		})
		return
	}
	c.JSON(http.StatusOK, models.JsonResult{
		Success: true,
		Error:   models.SUCCESS.Error(),
		Code:    models.SUCCESS,
		Data:    accountAssetTransfers,
	})
}

func GetAccountAssetBalanceLimitAndOffset(c *gin.Context) {
	var getAccountAssetBalanceLimitAndOffsetRequest services.GetAccountAssetBalanceLimitAndOffsetRequest
	err := c.ShouldBindJSON(&getAccountAssetBalanceLimitAndOffsetRequest)
	if err != nil {
		c.JSON(http.StatusOK, models.JsonResult{
			Success: false,
			Error:   err.Error(),
			Code:    models.ShouldBindJsonErr,
			Data:    nil,
		})
		return
	}
	assetId := getAccountAssetBalanceLimitAndOffsetRequest.AssetId
	limit := getAccountAssetBalanceLimitAndOffsetRequest.Limit
	offset := getAccountAssetBalanceLimitAndOffsetRequest.Offset

	{

		number, err := services.GetAccountAssetBalancePageNumberByPageSize(assetId, limit)

		pageNumber := offset/limit + 1
		if pageNumber > number {
			err = errors.New("page number must be greater than max value " + strconv.Itoa(number))
			c.JSON(http.StatusOK, models.JsonResult{
				Success: false,
				Error:   err.Error(),
				Code:    models.PageNumberExceedsTotalNumberErr,
				Data:    nil,
			})
			return
		}
	}

	accountAssetBalances, err := services.GetAccountAssetBalanceExtendsLimitAndOffset(assetId, limit, offset)
	if err != nil {
		c.JSON(http.StatusOK, models.JsonResult{
			Success: false,
			Error:   err.Error(),
			Code:    models.GetAccountAssetBalancesLimitAndOffsetErr,
			Data:    nil,
		})
		return
	}
	c.JSON(http.StatusOK, models.JsonResult{
		Success: true,
		Error:   models.SUCCESS.Error(),
		Code:    models.SUCCESS,
		Data:    accountAssetBalances,
	})
}

func GetAccountAssetBalancePageNumberByPageSize(c *gin.Context) {
	var getAccountAssetBalancePageNumberByPageSizeRequest services.GetAccountAssetBalancePageNumberByPageSizeRequest
	err := c.ShouldBindJSON(&getAccountAssetBalancePageNumberByPageSizeRequest)
	if err != nil {
		c.JSON(http.StatusOK, models.JsonResult{
			Success: false,
			Error:   err.Error(),
			Code:    models.ShouldBindJsonErr,
			Data:    nil,
		})
		return
	}
	pageSize := getAccountAssetBalancePageNumberByPageSizeRequest.PageSize
	assetId := getAccountAssetBalancePageNumberByPageSizeRequest.AssetId
	if pageSize <= 0 || assetId == "" {
		c.JSON(http.StatusOK, models.JsonResult{
			Success: false,
			Error:   errors.New("invalid asset id or page size").Error(),
			Code:    models.GetAccountAssetBalancePageNumberByPageSizeRequestInvalidErr,
			Data:    nil,
		})
		return
	}
	pageNumber, err := services.GetAccountAssetBalancePageNumberByPageSize(assetId, pageSize)
	if err != nil {
		c.JSON(http.StatusOK, models.JsonResult{
			Success: false,
			Error:   err.Error(),
			Code:    models.GetAccountAssetBalancePageNumberByPageSizeErr,
			Data:    nil,
		})
		return
	}
	c.JSON(http.StatusOK, models.JsonResult{
		Success: true,
		Error:   models.SuccessErr,
		Code:    models.SUCCESS,
		Data:    pageNumber,
	})
}

func GetAccountAssetTransferLimitAndOffset(c *gin.Context) {
	var getAccountAssetTransferLimitAndOffsetRequest services.GetAccountAssetTransferLimitAndOffsetRequest
	err := c.ShouldBindJSON(&getAccountAssetTransferLimitAndOffsetRequest)
	if err != nil {
		c.JSON(http.StatusOK, models.JsonResult{
			Success: false,
			Error:   err.Error(),
			Code:    models.ShouldBindJsonErr,
			Data:    nil,
		})
		return
	}
	assetId := getAccountAssetTransferLimitAndOffsetRequest.AssetId
	limit := getAccountAssetTransferLimitAndOffsetRequest.Limit
	offset := getAccountAssetTransferLimitAndOffsetRequest.Offset

	{

		number, err := services.GetAccountAssetTransferPageNumberByPageSize(assetId, limit)

		pageNumber := offset/limit + 1
		if pageNumber > number {
			err = errors.New("page number must be greater than max value " + strconv.Itoa(number))
			c.JSON(http.StatusOK, models.JsonResult{
				Success: false,
				Error:   err.Error(),
				Code:    models.PageNumberExceedsTotalNumberErr,
				Data:    nil,
			})
			return
		}
	}

	accountAssetTransfers, err := services.GetAccountAssetTransfersLimitAndOffset(assetId, limit, offset)
	if err != nil {
		c.JSON(http.StatusOK, models.JsonResult{
			Success: false,
			Error:   err.Error(),
			Code:    models.GetAccountAssetTransfersLimitAndOffsetErr,
			Data:    nil,
		})
		return
	}
	c.JSON(http.StatusOK, models.JsonResult{
		Success: true,
		Error:   models.SUCCESS.Error(),
		Code:    models.SUCCESS,
		Data:    accountAssetTransfers,
	})
}

func GetAccountAssetTransferPageNumberByPageSize(c *gin.Context) {
	var GetAccountAssetTransferPageNumberByPageSizeRequest services.GetAccountAssetTransferPageNumberByPageSizeRequest
	err := c.ShouldBindJSON(&GetAccountAssetTransferPageNumberByPageSizeRequest)
	if err != nil {
		c.JSON(http.StatusOK, models.JsonResult{
			Success: false,
			Error:   err.Error(),
			Code:    models.ShouldBindJsonErr,
			Data:    nil,
		})
		return
	}
	pageSize := GetAccountAssetTransferPageNumberByPageSizeRequest.PageSize
	assetId := GetAccountAssetTransferPageNumberByPageSizeRequest.AssetId
	if pageSize <= 0 || assetId == "" {
		c.JSON(http.StatusOK, models.JsonResult{
			Success: false,
			Error:   errors.New("invalid asset id or page size").Error(),
			Code:    models.GetAccountAssetTransferPageNumberByPageSizeRequestInvalidErr,
			Data:    nil,
		})
		return
	}
	pageNumber, err := services.GetAccountAssetTransferPageNumberByPageSize(assetId, pageSize)
	if err != nil {
		c.JSON(http.StatusOK, models.JsonResult{
			Success: false,
			Error:   err.Error(),
			Code:    models.GetAccountAssetTransferPageNumberByPageSizeErr,
			Data:    nil,
		})
		return
	}
	c.JSON(http.StatusOK, models.JsonResult{
		Success: true,
		Error:   models.SuccessErr,
		Code:    models.SUCCESS,
		Data:    pageNumber,
	})
}

func GetAllAccountAssetBalanceSimplified(c *gin.Context) {

}

func GetAccountAssetBalanceUserHoldTotalAmount(c *gin.Context) {
	assetId := c.Query("asset_id")
	totalAmount, err := services.GetAccountAssetBalanceUserHoldTotalAmount(assetId)
	if err != nil {
		c.JSON(http.StatusOK, models.JsonResult{
			Success: false,
			Error:   err.Error(),
			Code:    models.GetAccountAssetBalanceUserHoldTotalAmountErr,
			Data:    nil,
		})
		return
	}
	c.JSON(http.StatusOK, models.JsonResult{
		Success: true,
		Error:   models.SUCCESS.Error(),
		Code:    models.SUCCESS,
		Data:    totalAmount,
	})
}

func GetCustodyAssetRankList(c *gin.Context) {
	creds := struct {
		AssetId  string `json:"assetId"`
		Page     int    `json:"page"`
		PageSize int    `json:"pageSize"`
	}{}

	if err := c.ShouldBindJSON(&creds); err != nil {
		btlLog.CUST.Error("%v", err)
		c.JSON(http.StatusBadRequest, models.MakeJsonErrorResultForHttp(models.DefaultErr, "Json format error", nil))
		return
	}
	if creds.Page == 0 {
		c.JSON(http.StatusBadRequest, models.MakeJsonErrorResultForHttp(models.DefaultErr, "page must be greater than 0", nil))
		return
	}
	creds.Page = creds.Page - 1

	a, count, total := ManageQuery.GetAssetsBalanceRankList(creds.AssetId, creds.Page, creds.PageSize)
	list := struct {
		Count int64                               `json:"count"`
		List  *[]ManageQuery.GetAssetRankListResp `json:"list"`
		Total float64                             `json:"total"`
	}{
		Count: count,
		List:  a,
		Total: total,
	}
	c.JSON(http.StatusOK, models.MakeJsonErrorResultForHttp(models.SUCCESS, "", list))
}

func GetAccountAssetBalancePc(c *gin.Context) {

	var req struct {
		AssetId string `json:"asset_id"`
		Limit   int    `json:"limit"`
		Offset  int    `json:"offset"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		logrus.Errorln(errors.Wrap(err, "c.ShouldBindJSON"))
		c.JSON(http.StatusOK, models.RespLnc[*services.AccountAssetBalanceExtend]{
			Code: models.ToCode(models.ShouldBindJsonErr),
			Msg:  err.Error(),
			Data: models.LncT[*services.AccountAssetBalanceExtend]{
				List:  nil,
				Count: 0,
			},
		})
		return
	}

	if req.AssetId == "" || req.Limit < 0 || req.Offset < 0 {
		c.JSON(http.StatusOK, models.RespLnc[*services.AccountAssetBalanceExtend]{
			Code: models.ToCode(models.InvalidReq),
			Msg:  invalidReq.Error(),
			Data: models.LncT[*services.AccountAssetBalanceExtend]{
				List:  nil,
				Count: 0,
			},
		})
		return
	}

	count, err := services.GetAccountAssetBalanceCount(req.AssetId)
	if err != nil {
		c.JSON(http.StatusOK, models.RespLnc[*services.AccountAssetBalanceExtend]{
			Code: models.ToCode(models.GetAccountAssetBalanceCountErr),
			Msg:  err.Error(),
			Data: models.LncT[*services.AccountAssetBalanceExtend]{
				List:  nil,
				Count: 0,
			},
		})
		return
	}

	balances, err := services.GetAccountAssetBalance(req.AssetId, req.Limit, req.Offset)
	if err != nil {
		c.JSON(http.StatusOK, models.RespLnc[*services.AccountAssetBalanceExtend]{
			Code: models.ToCode(models.GetAccountAssetBalanceErr),
			Msg:  err.Error(),
			Data: models.LncT[*services.AccountAssetBalanceExtend]{
				List:  nil,
				Count: 0,
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.RespLnc[*services.AccountAssetBalanceExtend]{
		Code: models.ToCode(models.SUCCESS),
		Msg:  models.NullStr,
		Data: models.LncT[*services.AccountAssetBalanceExtend]{
			List:  balances,
			Count: count,
		},
	})
	return

}

func GetAccountAssetTransferPc(c *gin.Context) {

	var req struct {
		AssetId string `json:"asset_id"`
		Limit   int    `json:"limit"`
		Offset  int    `json:"offset"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		logrus.Errorln(errors.Wrap(err, "c.ShouldBindJSON"))
		c.JSON(http.StatusOK, models.RespLnc[*services.AccountAssetTransfer]{
			Code: models.ToCode(models.ShouldBindJsonErr),
			Msg:  err.Error(),
			Data: models.LncT[*services.AccountAssetTransfer]{
				List:  nil,
				Count: 0,
			},
		})
		return
	}

	if req.AssetId == "" || req.Limit < 0 || req.Offset < 0 {
		c.JSON(http.StatusOK, models.RespLnc[*services.AccountAssetTransfer]{
			Code: models.ToCode(models.InvalidReq),
			Msg:  invalidReq.Error(),
			Data: models.LncT[*services.AccountAssetTransfer]{
				List:  nil,
				Count: 0,
			},
		})
		return
	}

	count, err := services.GetAccountAssetTransferCount(req.AssetId)
	if err != nil {
		c.JSON(http.StatusOK, models.RespLnc[*services.AccountAssetTransfer]{
			Code: models.ToCode(models.GetAccountAssetBalanceCountErr),
			Msg:  err.Error(),
			Data: models.LncT[*services.AccountAssetTransfer]{
				List:  nil,
				Count: 0,
			},
		})
		return
	}

	transfers, err := services.GetAccountAssetTransfer(req.AssetId, req.Limit, req.Offset)
	if err != nil {
		c.JSON(http.StatusOK, models.RespLnc[*services.AccountAssetTransfer]{
			Code: models.ToCode(models.GetAccountAssetBalanceErr),
			Msg:  err.Error(),
			Data: models.LncT[*services.AccountAssetTransfer]{
				List:  nil,
				Count: 0,
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.RespLnc[*services.AccountAssetTransfer]{
		Code: models.ToCode(models.SUCCESS),
		Msg:  models.NullStr,
		Data: models.LncT[*services.AccountAssetTransfer]{
			List:  transfers,
			Count: count,
		},
	})
	return

}
