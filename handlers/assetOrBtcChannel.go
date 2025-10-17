package handlers

import (
	"net/http"
	"reflect"
	"trade/api"
	"trade/btlLog"
	"trade/models"
	"trade/services"
	"trade/services/custodyAccount/defaultAccount/custodyFee"
	"trade/utils"

	"github.com/gin-gonic/gin"
)

type BtcAssetRangeReq struct {
	MinSatsAmt     int `json:"min_sats_amt"`
	MaxSatsAmt     int `json:"max_sats_amt"`
	MinAssetAmount int `json:"min_asset_amount"`
	MaxAssetAmount int `json:"max_asset_amount"`
	MinLocalSats   int `json:"min_local_sats"`
	MaxLocalSats   int `json:"max_local_sats"`
}

func GetBtcAssetRange(c *gin.Context) {
	assetId := c.Param("asset_id")
	if assetId == "" {
		c.JSON(http.StatusOK, models.JsonResult{
			Success: false,
			Error:   "asset_id is empty",
			Data:    nil,
		})
		return
	}
	var response BtcAssetRangeReq
	chanAssetId, _ := services.ReadChannelAsset(assetId)
	if chanAssetId != nil {
		response = BtcAssetRangeReq{
			MinSatsAmt:     chanAssetId.MinSatsAmt,
			MaxSatsAmt:     chanAssetId.MaxSatsAmt,
			MinAssetAmount: chanAssetId.MinAssetAmount,
			MaxAssetAmount: chanAssetId.MaxAssetAmount,
			MinLocalSats:   chanAssetId.MinLocalSats,
			MaxLocalSats:   chanAssetId.MaxLocalSats,
		}
	} else {
		response = BtcAssetRangeReq{
			MinSatsAmt:     0,
			MaxSatsAmt:     0,
			MinAssetAmount: 0,
			MaxAssetAmount: 0,
			MinLocalSats:   0,
			MaxLocalSats:   0,
		}
	}

	c.JSON(http.StatusOK, models.JsonResult{
		Success: true,
		Error:   models.SuccessErr,
		Code:    models.SUCCESS,
		Data:    response,
	})
}

type ChannelFeeResp struct {
	Gasfee    int     `json:"gas_fee"`
	BaseFee   int     `json:"base_fee"`
	AssetRate float64 `json:"asset_rate"`
}

func GetOpenChannelFee(c *gin.Context) {
	assetId := c.Param("asset_id")
	if assetId == "" {
		c.JSON(http.StatusOK, models.JsonResult{
			Success: false,
			Error:   "asset_id is empty",
			Data:    nil,
		})
		return
	}

	var req ChannelFeeResp

	feeRate, err := services.GetMempoolFeeRate()
	if err != nil {
		c.JSON(http.StatusOK, models.JsonResult{
			Success: false,
			Error:   "Get FeeRate. " + err.Error(),
			Data:    nil,
		})
		return
	}

	req.Gasfee = feeRate.SatPerB.FastestFee * services.GetIssuanceTransactionByteSize()
	chanAssetId, _ := services.ReadChannelAsset(assetId)
	if chanAssetId != nil {
		req.BaseFee = chanAssetId.BaseFee
		req.AssetRate = chanAssetId.Rate
	} else {
		req.Gasfee = 0
		req.BaseFee = 0
		req.AssetRate = 0
	}

	c.JSON(http.StatusOK, models.JsonResult{
		Success: true,
		Error:   models.SuccessErr,
		Code:    models.SUCCESS,
		Data:    req,
	})
}

type TradeToUserFundChannelReq struct {
	AssetId    string `json:"asset_id"`
	Amount     int    `json:"amount"`
	PeerPubkey string `json:"peer_pubkey"`
	FeeRate    int    `json:"fee_rate"`
	PushSat    int    `json:"push_sat"`
	LocalSat   int    `json:"local_sat"`
}

func TradeToUserFundChannel(c *gin.Context) {
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
	fundReq := TradeToUserFundChannelReq{}
	if err := c.ShouldBindJSON(&fundReq); err != nil {
		c.JSON(http.StatusOK, models.MakeJsonErrorResultForHttp(models.DefaultErr, err.Error(), nil))
		return
	}
	btlLog.OpenChannel.Info("\nfundReq\n %v", utils.ValueJsonString(fundReq))

	str := services.ValidateFundChannelRequest(fundReq.AssetId, fundReq.Amount, fundReq.LocalSat)
	if str != "" {
		c.JSON(http.StatusOK, models.JsonResult{
			Success: false,
			Error:   str,
			Data:    nil,
		})
		return
	}
	feeRate, err := services.GetMempoolFeeRate()
	if err != nil {
		c.JSON(http.StatusOK, models.JsonResult{
			Success: false,
			Error:   "Get FeeRate. " + err.Error(),
			Data:    nil,
		})
		return
	}
	if feeRate.SatPerB.FastestFee < 3 || feeRate.SatPerB.FastestFee > 8 {
		feeRate.SatPerB.FastestFee = 3
	}
	var gasfee int
	if fundReq.AssetId == "00" {
		gasfee = feeRate.SatPerB.FastestFee*services.GetIssuanceTransactionByteSize() + fundReq.Amount/100
	} else {
		chanAssetId, _ := services.ReadChannelAsset(fundReq.AssetId)
		if chanAssetId == nil {
			c.JSON(http.StatusOK, models.JsonResult{
				Success: false,
				Error:   "not fund asset",
				Data:    nil,
			})
			return
		}
		fee := float64(fundReq.Amount) * chanAssetId.Rate
		gasfee = feeRate.SatPerB.FastestFee*services.GetIssuanceTransactionByteSize() + chanAssetId.BaseFee + int(fee) + fundReq.LocalSat/100
	}

	enough := custodyFee.IsAccountBalanceEnoughByUserId(uint(userId), uint64(gasfee))
	if !enough {
		c.JSON(http.StatusOK, models.JsonResult{
			Success: false,
			Error:   "Account Balance is not enough",
			Data:    nil,
		})
		return
	}

	peer, _ := api.GetChannelPeer(fundReq.PeerPubkey)
	if !peer {
		c.JSON(http.StatusOK, models.JsonResult{
			Success: false,
			Error:   "not fund peer",
			Data:    nil,
		})
		return
	}

	paidId, err := custodyFee.PayReverseChannelFee(uint(userId), uint64(gasfee))
	if err != nil {
		btlLog.OpenChannel.Error("\nUser %d PayGasFee Err %v \n", userId, err)
		c.JSON(http.StatusOK, models.JsonResult{
			Success: false,
			Error:   "PayReverseChannelFee Failed",
			Data:    nil,
		})
		return
	}

	record := &models.AssetOrBtcChannelRecord{
		AssetId:  fundReq.AssetId,
		Pubkey:   fundReq.PeerPubkey,
		Amount:   fundReq.Amount,
		LocalAmt: fundReq.LocalSat,
		PaidId:   int(paidId),
		UserId:   userId,
		Username: username,
	}

	err = services.CreateOpenChannelRecord(record)
	if err != nil {
		btlLog.OpenChannel.Error("\nUser %d Status 1 CreateOpenChannelRecord Err %v \n", userId, err)
	}

	channel, err := services.TradeToUserFundChannel(fundReq.AssetId, fundReq.PeerPubkey, fundReq.Amount, feeRate.SatPerB.FastestFee, fundReq.PushSat, fundReq.LocalSat)
	if err != nil {
		backId, err1 := custodyFee.BackReverseChannelFee(paidId)
		if err1 != nil {
			btlLog.OpenChannel.Error("\nUser %d CreateOpenChannelRecord Err %v \n", userId, err)
			record.Status = 4
			err2 := services.UpdateOpenChannelRecord(record)
			if err2 != nil {
				btlLog.OpenChannel.Error("\nUser %d Status 4 CreateOpenChannelRecord Err %v \n", userId, err)
			}
		} else {
			record.RefundsId = int(backId)
			record.Status = 3
			err2 := services.UpdateOpenChannelRecord(record)
			if err2 != nil {
				btlLog.OpenChannel.Error("\nUser %d Status 3 UpdateOpenChannelRecord Err %v \n", userId, err2)
			}
		}
		c.JSON(http.StatusOK, models.JsonResult{
			Success: false,
			Error:   err.Error(),
			Code:    models.TradeToUserFundChannelErr,
			Data:    nil,
		})
		return
	}
	record.ChannelPoint = channel
	record.Status = 2
	err = services.UpdateOpenChannelRecord(record)
	if err != nil {
		btlLog.OpenChannel.Error("\nUser %d Status 2 UpdateOpenChannelRecord Err %v \n", userId, err)
	}
	c.JSON(http.StatusOK, models.JsonResult{
		Success: true,
		Error:   models.SuccessErr,
		Code:    models.SUCCESS,
		Data:    channel,
	})
}

type GetAssetChannelListToChanidsReq struct {
	AssetId    string `json:"asset_id"`
	PeerPubkey string `json:"peer_pubkey"`
}

func GetChannelIdsAndPoints(c *gin.Context) {
	ChanReq := GetAssetChannelListToChanidsReq{}
	if err := c.ShouldBindJSON(&ChanReq); err != nil {
		c.JSON(http.StatusOK, models.MakeJsonErrorResultForHttp(models.DefaultErr, err.Error(), nil))
		return
	}
	btlLog.OpenChannel.Info("\nfundReq\n %v", utils.ValueJsonString(ChanReq))
	resp, err := api.GetChannelIdsAndPoints(ChanReq.AssetId, ChanReq.PeerPubkey)
	if err != nil {
		c.JSON(http.StatusOK, models.JsonResult{
			Success: false,
			Error:   err.Error(),
			Code:    models.GetChannelIdsAndPointsErr,
			Data:    nil,
		})
		return
	}
	c.JSON(http.StatusOK, models.JsonResult{
		Success: true,
		Error:   models.SuccessErr,
		Code:    models.SUCCESS,
		Data:    resp,
	})
}

type SetChannelAssetReq struct {
	AssetId        string  `json:"asset_id"`
	AssetName      string  `json:"asset_name"`
	MinSatsAmt     int     `json:"min_sats_amt"`
	MaxSatsAmt     int     `json:"max_sats_amt"`
	MinAssetAmount int     `json:"min_asset_amount"`
	MaxAssetAmount int     `json:"max_asset_amount"`
	MinLocalSats   int     `json:"min_local_sats"`
	MaxLocalSats   int     `json:"max_local_sats"`
	BaseFee        int     `json:"base_fee"`
	Rate           float64 `json:"rate"`
}

func SetChannelAsset(c *gin.Context) {
	req := SetChannelAssetReq{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, models.MakeJsonErrorResultForHttp(models.DefaultErr, err.Error(), nil))
		return
	}
	req1 := &models.ChannelAssets{
		AssetId:        req.AssetId,
		AssetName:      req.AssetName,
		MinSatsAmt:     req.MinSatsAmt,
		MaxSatsAmt:     req.MaxSatsAmt,
		MinAssetAmount: req.MinAssetAmount,
		MaxAssetAmount: req.MaxAssetAmount,
		MinLocalSats:   req.MinLocalSats,
		MaxLocalSats:   req.MaxLocalSats,
		BaseFee:        req.BaseFee,
		Rate:           req.Rate,
	}
	err := services.CreateChannelAsset(req1)
	if err != nil {
		c.JSON(http.StatusOK, models.JsonResult{
			Success: false,
			Error:   "Create ChannelAsset. " + err.Error(),
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

type UpdateChannelAssetReq struct {
	AssetId        *string  `json:"asset_id"`
	AssetName      *string  `json:"asset_name"`
	MinSatsAmt     *int     `json:"min_sats_amt"`
	MaxSatsAmt     *int     `json:"max_sats_amt"`
	MinAssetAmount *int     `json:"min_asset_amount"`
	MaxAssetAmount *int     `json:"max_asset_amount"`
	MinLocalSats   *int     `json:"min_local_sats"`
	MaxLocalSats   *int     `json:"max_local_sats"`
	BaseFee        *int     `json:"base_fee"`
	Rate           *float64 `json:"rate"`
}

func UpdateChannelAsset(c *gin.Context) {
	req := UpdateChannelAssetReq{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, models.MakeJsonErrorResultForHttp(models.DefaultErr, err.Error(), nil))
		return
	}
	updates := buildUpdateMap(&req)
	if len(updates) == 0 {
		c.JSON(http.StatusOK, models.JsonResult{
			Success: false,
			Error:   "no valid fields to update",
			Data:    nil,
		})
		return
	}
	err := services.UpdateChannelAsset(*req.AssetId, updates)
	if err != nil {
		c.JSON(http.StatusOK, models.JsonResult{
			Success: false,
			Error:   "Update ChannelAsset. " + err.Error(),
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

func buildUpdateMap(req *UpdateChannelAssetReq) map[string]any {
	updates := make(map[string]any)
	val := reflect.ValueOf(req).Elem()
	typ := val.Type()

	for i := range val.NumField() {
		field := typ.Field(i)
		fieldValue := val.Field(i)

		if fieldValue.Kind() == reflect.Ptr && !fieldValue.IsNil() {
			jsonTag := field.Tag.Get("json")
			if jsonTag != "" && jsonTag != "asset_id" {
				updates[jsonTag] = fieldValue.Elem().Interface()
			}
		}
	}
	return updates
}
