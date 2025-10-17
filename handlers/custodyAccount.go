package handlers

import (
	"fmt"
	"net/http"
	"trade/models"
	"trade/services/custodyAccount"
	"trade/services/custodyAccount/account"
	"trade/services/custodyAccount/custodyBase"
	"trade/services/custodyAccount/defaultAccount/custodyBtc"
	rpc "trade/services/servicesrpc"

	"github.com/gin-gonic/gin"
)

func ApplyInvoice(c *gin.Context) {

	userName := c.MustGet("username").(string)
	e, err := custodyBtc.NewBtcChannelEvent(userName)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error() + "用户不存在"})
		return
	}
	apply := custodyAccount.ApplyRequest{}
	if err = c.ShouldBindJSON(&apply); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	a := custodyBtc.BtcApplyInvoiceRequest{
		Amount: apply.Amount,
		Memo:   apply.Memo,
	}
	req, err := e.ApplyPayReq(&a)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"invoice": req.GetPayReq()})
}

func QueryInvoice(c *gin.Context) {

	userName := c.MustGet("username").(string)
	e, err := custodyBtc.NewBtcChannelEvent(userName)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error() + "用户不存在"})
		return
	}
	invoiceRequest := struct {
		AssetId string `json:"asset_id"`
	}{}
	if err := c.ShouldBindJSON(&invoiceRequest); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Request is erro"})
		return
	}

	invoices, err := e.QueryPayReq(invoiceRequest.AssetId)
	if err != nil {
		fmt.Println(err.Error())
		c.JSON(http.StatusUnauthorized, gin.H{"error": "service error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"invoices": invoices})
}

func PayInvoice(c *gin.Context) {

	userName := c.MustGet("username").(string)
	e, err := custodyBtc.NewBtcChannelEvent(userName)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error() + "用户不存在"})
		return
	}
	if !e.UserInfo.PayLock() {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "当前用户正在支付中，请勿频繁操作"})
		return
	}
	defer e.UserInfo.PayUnlock()

	pay := custodyAccount.PayInvoiceRequest{}
	if err := c.ShouldBindJSON(&pay); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error() + "请求参数错误"})
		return
	}
	a := custodyBtc.BtcPacket{
		PayReq: pay.Invoice,
	}
	err2 := e.SendPayment(&a)
	if err2 != nil {
		c.JSON(http.StatusOK, gin.H{"error": "SendPayment error:" + err2.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"payment": "success"})
}
func PayUserBtc(c *gin.Context) {

	userName := c.MustGet("username").(string)
	e, err := custodyBtc.NewBtcChannelEvent(userName)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error() + "用户不存在"})
		return
	}
	if !e.UserInfo.PayLock() {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "当前用户正在支付中，请勿频繁操作"})
		return
	}
	defer e.UserInfo.PayUnlock()

	pay := struct {
		NpubKey string  `json:"npub_key"`
		Amount  float64 `json:"amount"`
	}{}
	if err := c.ShouldBindJSON(&pay); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error() + "请求参数错误"})
		return
	}

	err = e.SendPaymentToUser(pay.NpubKey, pay.Amount)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"error": "SendPayment error:" + err.Error()})
		return
	}
	result := struct {
		Success string `json:"success"`
	}{
		Success: "success",
	}
	c.JSON(http.StatusOK, models.MakeJsonErrorResultForHttp(models.SUCCESS, "", result))
}

func PayBtcOnchain(c *gin.Context) {

	pay := struct {
		Address string  `json:"address"`
		Amount  float64 `json:"amount"`
	}{}
	if err := c.ShouldBindJSON(&pay); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error() + "请求参数错误"})
		return
	}

	userName := c.MustGet("username").(string)
	e, err := custodyBtc.NewBtcChannelEvent(userName)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error() + "用户不存在"})
		return
	}
	if !e.UserInfo.PayLock() {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "当前用户正在支付中，请勿频繁操作"})
		return
	}
	defer e.UserInfo.PayUnlock()

	err = e.PayToOutsideOnChain(pay.Address, pay.Amount)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"error": "SendPayment error:" + err.Error()})
		return
	}
	result := struct {
		Success string `json:"success"`
	}{
		Success: "success",
	}
	c.JSON(http.StatusOK, models.MakeJsonErrorResultForHttp(models.SUCCESS, "", result))
}

func QueryBalance(c *gin.Context) {

	userName := c.MustGet("username").(string)
	e, err := custodyBtc.NewBtcChannelEvent(userName)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error() + "用户不存在"})
		return
	}
	getBalance, err := e.GetBalance()
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"balance": getBalance[0].Amount})
}

func QueryPayment(c *gin.Context) {

	userName := c.MustGet("username").(string)
	e, err := custodyBtc.NewBtcChannelEvent(userName)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error() + "用户不存在"})
		return
	}

	query := custodyBase.PaymentRequest{}
	if err := c.ShouldBindJSON(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if query.AssetId != "00" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "asset_id类型错误"})
		return
	}
	if query.Page == 0 && query.PageSize == 0 {
		query.Page = 1
		query.PageSize = 1000
		query.Away = 5
	}
	p, err := e.GetTransactionHistory(&query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	p.Sort()

	c.JSON(http.StatusOK, gin.H{"payments": p.PaymentList})
}

func DecodeInvoice(c *gin.Context) {
	query := custodyAccount.DecodeInvoiceRequest{}
	if err := c.ShouldBindJSON(&query); err != nil {
		c.JSON(http.StatusOK, models.MakeJsonErrorResultForHttp(models.DefaultErr, "请求参数错误", nil))
		return
	}
	q, err := rpc.InvoiceDecode(query.Invoice)
	if err != nil {
		c.JSON(http.StatusOK, models.MakeJsonErrorResultForHttp(models.DefaultErr, "发票解析失败："+err.Error(), nil))
		return
	}
	result := struct {
		Amount    int64  `json:"amount"`
		Timestamp int64  `json:"timestamp"`
		Expiry    int64  `json:"expiry"`
		Memo      string `json:"memo"`
	}{
		Amount:    q.NumSatoshis,
		Timestamp: q.Timestamp,
		Expiry:    q.Expiry,
		Memo:      q.Description,
	}
	c.JSON(http.StatusOK, models.MakeJsonErrorResultForHttp(models.SUCCESS, "", result))
}

func GetRechargeBtcOnChainAddress(c *gin.Context) {
	address, err := custodyBtc.GetRechargeBtcAddress()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"error": "GetRechargeBtcAddress error:" + err.Error()})
		return
	}
	result := struct {
		Addr string `json:"addr"`
	}{
		Addr: address,
	}
	c.JSON(http.StatusOK, models.MakeJsonErrorResultForHttp(models.SUCCESS, "", result))
}

func RechargeBtcOnChain(c *gin.Context) {

	userName := c.MustGet("username").(string)
	usr, err := account.GetUserInfo(userName)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error() + "用户不存在"})
		return
	}

	recharge := struct {
		TxHash  string `json:"tx_hash"`
		Address string `json:"address"`
	}{}
	if err := c.ShouldBindJSON(&recharge); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error() + "请求参数错误"})
		return
	}

	err = custodyBtc.PutBtcOnChainMission(usr, recharge.TxHash, recharge.Address)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"error": "RechargeBtcOnChain error:" + err.Error()})
		return
	}
	result := struct {
		Success string `json:"success"`
	}{
		Success: "success",
	}
	c.JSON(http.StatusOK, models.MakeJsonErrorResultForHttp(models.SUCCESS, "", result))
}
