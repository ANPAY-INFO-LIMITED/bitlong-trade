package handlers

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"trade/btlLog"
	"trade/models"
	"trade/models/custodyModels/custodyswap"
	"trade/services/btldb"
	"trade/services/custodyAccount/custodyBase"
	"trade/services/custodyAccount/custodyBase/custodyFee"
	"trade/services/custodyAccount/defaultAccount/custodyAssets"
	"trade/services/custodyAccount/defaultAccount/custodyBtc/mempool"
	"trade/services/custodyAccount/defaultAccount/swap"
	rpc "trade/services/servicesrpc"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ApplyAddressRequest struct {
	Amount  float64 `json:"amount"`
	AssetId string  `json:"asset_id"`
}

func ApplyAddress(c *gin.Context) {

	userName := c.MustGet("username").(string)
	apply := ApplyAddressRequest{}
	if err := c.ShouldBindJSON(&apply); err != nil {
		c.JSON(http.StatusOK, models.MakeJsonErrorResultForHttp(models.DefaultErr, err.Error(), nil))
		return
	}
	e, err := custodyAssets.NewAssetEvent(userName, apply.AssetId)
	if err != nil {
		c.JSON(http.StatusOK, models.MakeJsonErrorResultForHttp(models.DefaultErr, err.Error(), nil))
		return
	}
	req, err := e.ApplyPayReq(&custodyAssets.AssetAddressApplyRequest{
		Amount: int64(apply.Amount),
	})
	if err != nil {
		c.JSON(http.StatusOK, models.MakeJsonErrorResultForHttp(models.DefaultErr, err.Error(), nil))
		return
	}
	addr := struct {
		Address string `json:"addr"`
	}{
		Address: req.GetPayReq(),
	}
	c.JSON(http.StatusOK, models.MakeJsonErrorResultForHttp(models.SUCCESS, "", addr))
}

type ApplyAssetInvoiceRequest struct {
	Amount  float64 `json:"amount"`
	AssetId string  `json:"asset_id"`
}

func ApplyAssetInvoice(c *gin.Context) {

	userName := c.MustGet("username").(string)
	apply := ApplyAssetInvoiceRequest{}
	if err := c.ShouldBindJSON(&apply); err != nil {
		c.JSON(http.StatusOK, models.MakeJsonErrorResultForHttp(models.DefaultErr, err.Error(), nil))
		return
	}
	e, err := custodyAssets.NewAssetEvent(userName, apply.AssetId)
	if err != nil {
		c.JSON(http.StatusOK, models.MakeJsonErrorResultForHttp(models.DefaultErr, err.Error(), nil))
		return
	}
	req, err := e.ApplyChannelPayReq(&custodyAssets.AssetInvoiceApplyResponse{
		Amount: int64(apply.Amount),
	})
	if err != nil {
		c.JSON(http.StatusOK, models.MakeJsonErrorResultForHttp(models.DefaultErr, err.Error(), nil))
		return
	}
	Invoice := struct {
		Invoice string `json:"invoice"`
	}{
		Invoice: req.GetPayReq(),
	}
	c.JSON(http.StatusOK, models.MakeJsonErrorResultForHttp(models.SUCCESS, "", Invoice))
}

type SendAssetRequest struct {
	Address  string `json:"address"`
	AssetId  string `json:"asset_id"`
	PeerNode string `json:"rfq_peer_key"`
}

func SendAsset(c *gin.Context) {

	userName := c.MustGet("username").(string)
	req := SendAssetRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, models.MakeJsonErrorResultForHttp(models.DefaultErr, err.Error(), nil))
		return
	}
	e, err := custodyAssets.NewAssetEvent(userName, req.AssetId)
	if err != nil {
		c.JSON(http.StatusOK, models.MakeJsonErrorResultForHttp(models.DefaultErr, err.Error(), nil))
		return
	}

	if !e.UserInfo.PayLock() {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "当前用户正在支付中，请勿频繁操作"})
		return
	}
	defer e.UserInfo.PayUnlock()

	err = e.SendPayment(&custodyAssets.AssetPacket{
		PayReq: req.Address,
	})
	if err != nil {
		c.JSON(http.StatusOK, models.MakeJsonErrorResultForHttp(models.DefaultErr, err.Error(), nil))
		return
	}
	result := struct {
		Success string `json:"success"`
	}{
		Success: "success",
	}
	c.JSON(http.StatusOK, models.MakeJsonErrorResultForHttp(models.SUCCESS, "", result))
}

func SendToUserAsset(c *gin.Context) {

	userName := c.MustGet("username").(string)
	e, err := custodyAssets.NewAssetEvent(userName, "")
	if err != nil {
		c.JSON(http.StatusOK, models.MakeJsonErrorResultForHttp(models.DefaultErr, err.Error(), nil))
		return
	}

	if !e.UserInfo.PayLock() {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "当前用户正在支付中，请勿频繁操作"})
		return
	}
	defer e.UserInfo.PayUnlock()

	pay := struct {
		NpubKey string  `json:"npub_key"`
		AssetId string  `json:"asset_id"`
		Amount  float64 `json:"amount"`
	}{}
	if err = c.ShouldBindJSON(&pay); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error() + "请求参数错误"})
		return
	}

	err = e.SendPaymentToUser(pay.NpubKey, pay.Amount, pay.AssetId)
	if err != nil {
		c.JSON(http.StatusOK, models.MakeJsonErrorResultForHttp(models.DefaultErr, err.Error(), nil))
		return
	}
	result := struct {
		Success string `json:"success"`
	}{
		Success: "success",
	}
	c.JSON(http.StatusOK, models.MakeJsonErrorResultForHttp(models.SUCCESS, "", result))
}

type QueryAssetRequest struct {
	AssetId string `json:"asset_id"`
}

func QueryAsset(c *gin.Context) {
	userName := c.MustGet("username").(string)
	apply := QueryAssetRequest{}
	if err := c.ShouldBindJSON(&apply); err != nil {
		c.JSON(http.StatusOK, models.MakeJsonErrorResultForHttp(models.DefaultErr, err.Error(), nil))
		return
	}
	e, err := custodyAssets.NewAssetEvent(userName, apply.AssetId)
	if err != nil {
		c.JSON(http.StatusOK, models.MakeJsonErrorResultForHttp(models.DefaultErr, err.Error(), nil))
		return
	}
	balance, err := e.GetBalance()
	if err != nil {
		c.JSON(http.StatusOK, models.MakeJsonErrorResultForHttp(models.DefaultErr, err.Error(), nil))
		return
	}
	c.JSON(http.StatusOK, models.MakeJsonErrorResultForHttp(models.SUCCESS, "", balance))
}

func QueryAssets(c *gin.Context) {
	userName := c.MustGet("username").(string)
	e, err := custodyAssets.NewAssetEvent(userName, "")
	if err != nil {
		c.JSON(http.StatusOK, models.MakeJsonErrorResultForHttp(models.DefaultErr, err.Error(), nil))
		return
	}
	balance, err := e.GetBalances()
	if err != nil {
		c.JSON(http.StatusOK, models.MakeJsonErrorResultForHttp(models.DefaultErr, err.Error(), nil))
		return
	}
	request := DealBalance(balance)
	if request == nil {
		c.JSON(http.StatusOK, models.MakeJsonErrorResultForHttp(models.SUCCESS, "", balance))
	} else {
		c.JSON(http.StatusOK, models.MakeJsonErrorResultForHttp(models.SUCCESS, "", request))
	}
}

type AddressResponce struct {
	Address    string  `json:"addr"`
	AssetId    string  `json:"asset_id"`
	Amount     float64 `json:"amount"`
	CreateTime int64   `json:"createTime"`
}

func QueryAddress(c *gin.Context) {

	userName := c.MustGet("username").(string)
	invoiceRequest := struct {
		AssetId string `json:"asset_id"`
	}{}
	if err := c.ShouldBindJSON(&invoiceRequest); err != nil {
		c.JSON(http.StatusOK, models.MakeJsonErrorResultForHttp(models.DefaultErr, err.Error(), nil))
		return
	}
	e, err := custodyAssets.NewAssetEvent(userName, invoiceRequest.AssetId)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.MakeJsonErrorResultForHttp(models.DefaultErr, "用户不存在", nil))
		return
	}

	addr, err := e.QueryPayReq()
	if err != nil {
		fmt.Println(err.Error())
		c.JSON(http.StatusOK, models.MakeJsonErrorResultForHttp(models.DefaultErr, err.Error(), nil))
		return
	}
	var addrs []AddressResponce
	for _, v := range addr {
		addrs = append(addrs, AddressResponce{
			Address:    v.Invoice,
			AssetId:    v.AssetId,
			Amount:     v.Amount,
			CreateTime: v.CreatedAt.Unix(),
		})
	}
	c.JSON(http.StatusOK, models.MakeJsonErrorResultForHttp(models.SUCCESS, "", addrs))
}

func QueryAddresses(c *gin.Context) {

	userName := c.MustGet("username").(string)
	e, err := custodyAssets.NewAssetEvent(userName, "")
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.MakeJsonErrorResultForHttp(models.DefaultErr, "用户不存在", nil))
		return
	}

	addr, err := e.QueryPayReqs()
	if err != nil {
		fmt.Println(err.Error())
		c.JSON(http.StatusOK, models.MakeJsonErrorResultForHttp(models.DefaultErr, err.Error(), nil))
		return
	}
	var addrs []AddressResponce
	for _, v := range addr {
		addrs = append(addrs, AddressResponce{
			Address: v.Invoice,
			AssetId: v.AssetId,
			Amount:  v.Amount,
		})
	}
	c.JSON(http.StatusOK, models.MakeJsonErrorResultForHttp(models.SUCCESS, "", addrs))
}

type InvoiceResponce struct {
	Invoice string               `json:"invoice"`
	AssetId string               `json:"asset_id"`
	Amount  int64                `json:"amount"`
	Status  models.InvoiceStatus `json:"status"`
}

func QueryAssetInvoice(c *gin.Context) {

	userName := c.MustGet("username").(string)
	invoiceRequest := struct {
		AssetId string `json:"asset_id"`
	}{}
	if err := c.ShouldBindJSON(&invoiceRequest); err != nil {
		c.JSON(http.StatusOK, models.MakeJsonErrorResultForHttp(models.DefaultErr, err.Error(), nil))
		return
	}
	e, err := custodyAssets.NewAssetEvent(userName, invoiceRequest.AssetId)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.MakeJsonErrorResultForHttp(models.DefaultErr, "用户不存在", nil))
		return
	}

	resp, err := e.QueryChannelPayReq()
	if err != nil {
		fmt.Println(err.Error())
		c.JSON(http.StatusOK, models.MakeJsonErrorResultForHttp(models.DefaultErr, err.Error(), nil))
		return
	}
	var invoices []InvoiceResponce
	for _, v := range resp {
		invoices = append(invoices, InvoiceResponce{
			Invoice: v.Invoice,
			AssetId: v.AssetId,
			Amount:  int64(v.Amount),
			Status:  v.Status,
		})
	}
	c.JSON(http.StatusOK, models.MakeJsonErrorResultForHttp(models.SUCCESS, "", invoices))
}

func QueryAssetPayments(c *gin.Context) {

	userName := c.MustGet("username").(string)
	e, err := custodyAssets.NewAssetEvent(userName, "")
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.MakeJsonErrorResultForHttp(models.DefaultErr, "用户不存在", nil))
		return
	}
	transferRequest := custodyBase.PaymentRequest{}
	if err := c.ShouldBindJSON(&transferRequest); err != nil {
		c.JSON(http.StatusOK, models.MakeJsonErrorResultForHttp(models.DefaultErr, err.Error(), nil))
		return
	}
	if transferRequest.Page == 0 && transferRequest.PageSize == 0 {
		transferRequest.Page = 1
		transferRequest.PageSize = 1000
		transferRequest.Away = 5
	}

	payments, err := e.GetTransactionHistory(&transferRequest)
	if err != nil {
		fmt.Println(err.Error())
		c.JSON(http.StatusOK, models.MakeJsonErrorResultForHttp(models.DefaultErr, err.Error(), nil))
		return
	}
	c.JSON(http.StatusOK, models.MakeJsonErrorResultForHttp(models.SUCCESS, "", payments))
}

type AssetBalance struct {
	AssetId string  `json:"assetId"`
	Amount  int64   `json:"amount"`
	Price   float64 `json:"prices"`
}

func DealBalance(b []custodyBase.Balance) *[]AssetBalance {
	baseURL := "http:
	queryParams := url.Values{}
	t := make(map[string]int64)
	for _, v := range b {
		if v.AssetId == "00" {
			queryParams.Add("ids", "btc")
		} else {
			queryParams.Add("ids", v.AssetId)
		}
		queryParams.Add("numbers", strconv.FormatInt(v.Amount, 10))
		t[v.AssetId] = v.Amount
	}
	reqURL := baseURL + "?" + queryParams.Encode()

	resp, err := http.Get(reqURL)
	if err != nil {
		btlLog.CUST.Error("Error making request:", err)
		return nil
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			btlLog.CUST.Error("Error closing response body:", err)
		}
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		btlLog.CUST.Error("Error reading response body:", err)
		return nil
	}
	type temp struct {
		AssetsId string  `json:"id"`
		Price    float64 `json:"price"`
	}
	type List struct {
		List []temp `json:"list"`
	}
	r := struct {
		Success bool   `json:"success"`
		Error   string `json:"error"`
		Code    int    `json:"code"`
		Data    List   `json:"data"`
	}{}
	err = json.Unmarshal(body, &r)
	if err != nil {
		btlLog.CUST.Error(err.Error())
		return nil
	}
	var list []AssetBalance
	for _, v := range r.Data.List {
		if v.AssetsId == "btc" {
			v.AssetsId = "00"
		}
		list = append(list, AssetBalance{
			AssetId: v.AssetsId,
			Amount:  t[v.AssetsId],
			Price:   v.Price,
		})
	}
	return &list
}

type DecodeAddressRequest struct {
	Address string `json:"addr"`
}

func DecodeAddress(c *gin.Context) {
	query := DecodeAddressRequest{}
	if err := c.ShouldBindJSON(&query); err != nil {
		c.JSON(http.StatusOK, models.MakeJsonErrorResultForHttp(models.DefaultErr, "请求参数错误", nil))
		return
	}
	if query.Address == "" {
		c.JSON(http.StatusOK, models.MakeJsonErrorResultForHttp(models.DefaultErr, "请求参数错误", nil))
	}
	q, err := rpc.DecodeAddr(query.Address)
	if err != nil {
		c.JSON(http.StatusOK, models.MakeJsonErrorResultForHttp(models.DefaultErr, "地址解析失败："+err.Error(), nil))
		return
	}
	AssetId := hex.EncodeToString(q.AssetId)
	result := struct {
		AssetId   string  `json:"AssetId"`
		AssetType string  `json:"assetType"`
		Amount    uint64  `json:"amount"`
		FeeRate   float64 `json:"feeRate"`
	}{
		AssetId:   AssetId,
		AssetType: q.AssetType.String(),
		Amount:    q.Amount,
	}

	_, err = btldb.GetInvoiceByReq(query.Address)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		result.FeeRate = 0
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		result.FeeRate = float64(mempool.GetCustodyAssetFee())
	} else {
		result.FeeRate = float64(custodyFee.AssetInsideFee)
	}
	c.JSON(http.StatusOK, models.MakeJsonErrorResultForHttp(models.SUCCESS, "", result))
}

type DecodeInvoiceRequest struct {
	Invoice  string `json:"invoice"`
	AssetId  string `json:"asset_id"`
	PeerNode string `json:"rfq_peer_key"`
}

func DecodeAssetInvoice(c *gin.Context) {
	query := DecodeInvoiceRequest{}
	if err := c.ShouldBindJSON(&query); err != nil {
		c.JSON(http.StatusOK, models.MakeJsonErrorResultForHttp(models.DefaultErr, "请求参数错误", nil))
		return
	}
	if len(query.Invoice) < 10 || query.AssetId == "" {
		c.JSON(http.StatusOK, models.MakeJsonErrorResultForHttp(models.DefaultErr, "请求参数错误", nil))
	}
	if query.Invoice[0:2] != "ln" {
		c.JSON(http.StatusOK, models.MakeJsonErrorResultForHttp(models.DefaultErr, "非资产通道发票", nil))
	}

	result := struct {
		AssetId string  `json:"AssetId"`
		Amount  uint64  `json:"amount"`
		FeeRate float64 `json:"feeRate"`
	}{
		AssetId: query.AssetId,
	}

	invoice, err := btldb.GetInvoiceByReq(query.Invoice)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusOK, models.MakeJsonErrorResultForHttp(models.DefaultErr, "解析发票失败："+err.Error(), nil))
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		d, err := rpc.DecodeAssetInvoice(query.Invoice, query.AssetId)
		if err != nil {
			c.JSON(http.StatusOK, models.MakeJsonErrorResultForHttp(models.DefaultErr, "解析发票失败："+err.Error(), nil))
			return
		}
		result.Amount = d.AssetAmount
		c.JSON(http.StatusOK, models.MakeJsonErrorResultForHttp(models.SUCCESS, "", result))
		return
	} else {
		if invoice.AssetId == query.AssetId {
			result.Amount = uint64(invoice.Amount)
			c.JSON(http.StatusOK, models.MakeJsonErrorResultForHttp(models.SUCCESS, "", result))
		} else {
			c.JSON(http.StatusOK, models.MakeJsonErrorResultForHttp(models.DefaultErr, fmt.Sprintf("该发票仅支持使用资产ID(%s)支付",
				query.AssetId), nil))
		}
	}
}

type SetReceiveAssetRequest struct {
	AssetId string `json:"asset_id"`
	Enable  bool   `json:"enable"`
}

func AddReceiveAsset(c *gin.Context) {
	userName := c.MustGet("username").(string)
	req := SetReceiveAssetRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, models.MakeJsonErrorResultForHttp(models.DefaultErr, "请求参数错误", nil))
		return
	}
	if len(req.AssetId) < 10 {
		c.JSON(http.StatusOK, models.MakeJsonErrorResultForHttp(models.DefaultErr, "assetID格式请求参数错误", nil))
		return
	}
	if req.Enable {
		err := swap.AddReceiveConfig(userName, req.AssetId)
		if err != nil {
			c.JSON(http.StatusOK, models.MakeJsonErrorResultForHttp(models.DefaultErr, "设置失败："+err.Error(), nil))
			return
		}
		c.JSON(http.StatusOK, models.MakeJsonErrorResultForHttp(models.SUCCESS, "", nil))
	} else {
		err := swap.DeleteReceiveConfig(userName, req.AssetId)
		if err != nil {
			c.JSON(http.StatusOK, models.MakeJsonErrorResultForHttp(models.DefaultErr, "设置失败："+err.Error(), nil))
			return
		}
		c.JSON(http.StatusOK, models.MakeJsonErrorResultForHttp(models.SUCCESS, "", nil))
	}

}

type CheckReceiveAssetRequest struct {
	AssetId string `json:"asset_id"`
}

func CheckReceiveAsset(c *gin.Context) {
	userName := c.MustGet("username").(string)
	req := SetReceiveAssetRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, models.MakeJsonErrorResultForHttp(models.DefaultErr, "请求参数错误", nil))
		return
	}
	if len(req.AssetId) < 10 {
		c.JSON(http.StatusOK, models.MakeJsonErrorResultForHttp(models.DefaultErr, "assetID格式请求参数错误", nil))
		return
	}
	enable, err := swap.CheckReceiveConfig(userName, req.AssetId)
	if err != nil {
		c.JSON(http.StatusOK, models.MakeJsonErrorResultForHttp(models.DefaultErr, "获取失败："+err.Error(), nil))
		return
	}
	c.JSON(http.StatusOK, models.MakeJsonErrorResultForHttp(models.SUCCESS, "", enable))
}

func GetReceiveAssetSort(c *gin.Context) {
	userName := c.MustGet("username").(string)
	assetList, err := swap.GetReceiveSort(userName)
	if err != nil {
		c.JSON(http.StatusOK, models.MakeJsonErrorResultForHttp(models.DefaultErr, "获取失败："+err.Error(), nil))
		return
	}
	c.JSON(http.StatusOK, models.MakeJsonErrorResultForHttp(models.SUCCESS, "", assetList.AssetList))
}

type UpdateReceiveAssetSortRequest struct {
	AssetList []string `json:"asset_list"`
}

func UpdateReceiveAssetSort(c *gin.Context) {
	userName := c.MustGet("username").(string)
	req := UpdateReceiveAssetSortRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, models.MakeJsonErrorResultForHttp(models.DefaultErr, "请求参数错误", nil))
		return
	}
	err := swap.UpdateReceiveSort(userName, req.AssetList)
	if err != nil {
		c.JSON(http.StatusOK, models.MakeJsonErrorResultForHttp(models.DefaultErr, "更新失败："+err.Error(), nil))
		return
	}
	c.JSON(http.StatusOK, models.MakeJsonErrorResultForHttp(models.SUCCESS, "", true))
}

func GetCharacter(c *gin.Context) {
	userName := c.MustGet("username").(string)
	rep, err := swap.GetCharacter(userName)
	if err != nil {
		c.JSON(http.StatusOK, models.MakeJsonErrorResultForHttp(models.DefaultErr, "获取失败："+err.Error(), nil))
		return
	}
	var character string
	if rep == 0 {
		character = "customer"
	} else {
		character = "merchant"
	}
	c.JSON(http.StatusOK, models.MakeJsonErrorResultForHttp(models.SUCCESS, "", character))
}

type SetCharacterRequest struct {
	Character string `json:"character"`
}

func SetCharacter(c *gin.Context) {
	userName := c.MustGet("username").(string)
	req := SetCharacterRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, models.MakeJsonErrorResultForHttp(models.DefaultErr, "请求参数错误", nil))
		return
	}
	if req.Character != "customer" && req.Character != "merchant" {
		c.JSON(http.StatusOK, models.MakeJsonErrorResultForHttp(models.DefaultErr, "请求参数错误", nil))
		return
	}
	var character custodyswap.ReceiveCharacter
	if req.Character == "merchant" {
		character = custodyswap.ReceiveCharacterproducer
	}
	err := swap.SetCharacter(userName, character)
	if err != nil {
		c.JSON(http.StatusOK, models.MakeJsonErrorResultForHttp(models.DefaultErr, "设置失败："+err.Error(), nil))
		return
	}
	c.JSON(http.StatusOK, models.MakeJsonErrorResultForHttp(models.SUCCESS, "", true))
}
