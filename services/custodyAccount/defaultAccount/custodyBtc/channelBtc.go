package custodyBtc

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
	"trade/btlLog"
	"trade/middleware"
	"trade/models"
	"trade/models/custodyModels"
	"trade/services/btldb"
	caccount "trade/services/custodyAccount/account"
	cBase "trade/services/custodyAccount/custodyBase"
	"trade/services/custodyAccount/custodyBase/control"
	"trade/services/custodyAccount/custodyBase/custodyFee"
	"trade/services/custodyAccount/custodyBase/custodyLimit"
	"trade/services/custodyAccount/custodyBase/custodyPayTN"
	"trade/services/custodyAccount/defaultAccount/custodyBalance"
	rpc "trade/services/servicesrpc"

	"gorm.io/gorm"
)

type BtcChannelEvent struct {
	UserInfo *caccount.UserInfo
}

func NewBtcChannelEvent(UserName string) (*BtcChannelEvent, error) {
	var (
		e   BtcChannelEvent
		err error
	)
	e.UserInfo, err = caccount.GetUserInfo(UserName)
	if err != nil {
		btlLog.CUST.Warning("%s,UserName:%s", err.Error(), UserName)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: %s", caccount.CustodyAccountGetErr, "userName不存在")
		}
		return nil, fmt.Errorf("%w: %w", caccount.CustodyAccountGetErr, err)
	}
	btlLog.CUST.Info("UserName:%s", UserName)
	return &e, nil
}

func NewBtcChannelEventByUserId(UserId uint) (*BtcChannelEvent, error) {
	var (
		e   BtcChannelEvent
		err error
	)
	e.UserInfo, err = caccount.GetUserInfoById(UserId)
	if err != nil {
		btlLog.CUST.Error("%s,UserName:%s", err.Error(), UserId)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: %s", caccount.CustodyAccountGetErr, "userName不存在")
		}
		return nil, fmt.Errorf("%w: %w", caccount.CustodyAccountGetErr, err)
	}
	btlLog.CUST.Info("UserName:%s", UserId)
	return &e, nil
}

func (e *BtcChannelEvent) GetBalance() ([]cBase.Balance, error) {
	DB := middleware.DB
	balance := custodyBalance.GetBtcBalance(DB, e.UserInfo.Account.ID)
	balances := []cBase.Balance{
		{
			AssetId: "00",
			Amount:  int64(balance),
		},
	}
	return balances, nil
}

var CreateInvoiceErr = errors.New("CreateInvoiceErr")

func (e *BtcChannelEvent) ApplyPayReq(applyRequest *BtcApplyInvoiceRequest) (*BtcApplyInvoice, error) {
	if false {
		return nil, errors.New("测试期间，发票申请关闭")
	}

	invoice, err := rpc.InvoiceCreate(applyRequest.Amount, applyRequest.Memo)
	if err != nil {
		btlLog.CUST.Error(err.Error())
		return nil, fmt.Errorf("%w: %s", CreateInvoiceErr, err.Error())
	}

	info, _ := rpc.InvoiceFind(invoice.RHash)

	var invoiceModel models.Invoice
	invoiceModel.UserID = e.UserInfo.User.ID
	invoiceModel.Invoice = invoice.PaymentRequest
	invoiceModel.AccountID = &e.UserInfo.Account.ID
	invoiceModel.Amount = float64(info.Value)

	invoiceModel.Status = models.InvoiceStatus(info.State)
	template := time.Unix(info.CreationDate, 0)
	invoiceModel.CreateDate = &template
	expiry := int(info.Expiry)
	invoiceModel.Expiry = &expiry

	err = btldb.CreateInvoice(&invoiceModel)
	if err != nil {
		btlLog.CUST.Error(err.Error(), models.ReadDbErr)
		return nil, models.ReadDbErr
	}
	return &BtcApplyInvoice{
		LnInvoice: invoice,
		Amount:    applyRequest.Amount,
	}, nil
}

func (e *BtcChannelEvent) QueryPayReq(assetId string) ([]InvoiceResponce, error) {
	if assetId == "" {
		assetId = "00"
	}
	params := btldb.QueryParams{
		"UserID":  e.UserInfo.User.ID,
		"AssetId": assetId,
	}
	a, err := btldb.GenericQuery(&models.Invoice{}, params)
	if err != nil {
		btlLog.CUST.Error(err.Error())
		return nil, err
	}
	if len(a) > 0 {
		var invoices []InvoiceResponce
		for j := len(a) - 1; j >= 0; j-- {
			var i InvoiceResponce
			i.Invoice = a[j].Invoice
			i.AssetId = a[j].AssetId
			i.Amount = int64(a[j].Amount)
			i.Status = a[j].Status
			i.Time = a[j].CreatedAt.Unix()
			invoices = append(invoices, i)
		}
		return invoices, nil
	}
	return nil, nil
}

func (e *BtcChannelEvent) SendPayment(bt *BtcPacket) error {
	bt.err = make(chan error, 1)

	err := bt.VerifyPayReq(e.UserInfo)
	if err != nil {
		return err
	}
	if bt.isInsideMission != nil {
		if !control.GetTransferControl("00", control.TransferControlLocal) {
			return errors.New("当前服务调用失败，请稍后再试")
		}

		bt.isInsideMission.err = bt.err
		go e.payToInside(bt)
	} else {
		if !control.GetTransferControl("00", control.TransferControlOffChain) {
			return errors.New("当前服务调用失败，请稍后再试")
		}

		go e.payToOutside(bt)
	}
	ctx, cancel := context.WithTimeout(context.Background(), cBase.Timeout)
	defer cancel()
	select {
	case <-ctx.Done():
		go func(c chan error) {
			err := <-c
			if err != nil {
				btlLog.CUST.Error("btc sendPayment timeout:%s", err.Error())
			}
			close(c)
		}(bt.err)

		return cBase.TimeoutErr
	case err := <-bt.err:

		return err
	}
}

func (e *BtcChannelEvent) SendPaymentToUser(receiverUserName string, amount float64) error {
	if !control.GetTransferControl("00", control.TransferControlLocal) {
		return errors.New("当前服务调用失败，请稍后再试")
	}

	var err error
	receiver, err := caccount.GetUserInfo(receiverUserName)
	if err != nil {
		btlLog.CUST.Warning("%s,UserName:%s", err.Error(), receiverUserName)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("%w: %s", caccount.CustodyAccountGetErr, "userName不存在")
		}
		return fmt.Errorf("%w: %w", caccount.CustodyAccountGetErr, err)
	}

	limitType := custodyModels.LimitType{
		AssetId:      "00",
		TransferType: custodyModels.LimitTransferTypeLocal,
	}
	err = custodyLimit.CheckLimit(middleware.DB, e.UserInfo, &limitType, amount)
	if err != nil {
		return err
	}
	if !custodyBalance.CheckBtcBalance(middleware.DB, e.UserInfo, amount) {
		return NotSufficientFunds
	}

	m := custodyModels.AccountInsideMission{
		AccountId:  e.UserInfo.Account.ID,
		AssetId:    custodyBalance.BtcId,
		Type:       custodyModels.AIMTypeBtc,
		ReceiverId: receiver.Account.ID,
		InvoiceId:  0,
		Amount:     amount,
		Fee:        float64(custodyFee.ChannelBtcInsideServiceFee),
		FeeType:    custodyBalance.BtcId,
		State:      custodyModels.AIMStatePending,
	}
	LogAIM(middleware.DB, &m)
	err = RunInsidePTNStep(e.UserInfo, receiver, &m)
	if err != nil {
		return err
	}
	return nil
}

func (e *BtcChannelEvent) payToInside(bt *BtcPacket) {
	m := custodyModels.AccountInsideMission{
		AccountId:  e.UserInfo.Account.ID,
		AssetId:    custodyBalance.BtcId,
		Type:       custodyModels.AIMTypeBtc,
		ReceiverId: *bt.isInsideMission.insideInvoice.AccountID,
		InvoiceId:  bt.isInsideMission.insideInvoice.ID,
		Amount:     float64(bt.DecodePayReq.NumSatoshis),
		Fee:        float64(custodyFee.ChannelBtcInsideServiceFee),
		FeeType:    custodyBalance.BtcId,
		State:      custodyModels.AIMStatePending,
	}
	LogAIM(middleware.DB, &m)
	err := RunInsideStep(e.UserInfo, &m)
	bt.err <- err
}

func (e *BtcChannelEvent) payToOutside(bt *BtcPacket) {
	tx, back := middleware.GetTx()
	defer back()

	var balanceModel models.Balance
	balanceModel.AccountId = e.UserInfo.Account.ID
	balanceModel.BillType = models.BillTypePayment
	balanceModel.Away = models.AWAY_OUT
	balanceModel.Amount = float64(bt.DecodePayReq.NumSatoshis)
	balanceModel.Unit = models.UNIT_SATOSHIS
	balanceModel.Invoice = &bt.PayReq
	balanceModel.PaymentHash = &bt.DecodePayReq.PaymentHash
	balanceModel.State = models.STATE_UNKNOW
	balanceModel.TypeExt = &models.BalanceTypeExt{Type: models.BTExtOnChannel}
	err := btldb.CreateBalance(tx, &balanceModel)
	if err != nil {
		btlLog.CUST.Error(err.Error())
	}

	outsideMission := custodyModels.AccountOutsideMission{
		AccountId: e.UserInfo.Account.ID,
		AssetId:   "00",
		Type:      custodyModels.AOMTypeBtc,
		Target:    bt.PayReq,
		Hash:      bt.DecodePayReq.PaymentHash,
		Amount:    float64(bt.DecodePayReq.NumSatoshis),
		FeeLimit:  float64(bt.FeeLimit),
		BalanceId: balanceModel.ID,
		State:     custodyModels.AOMStatePending,
	}
	LogAOM(tx, &outsideMission)
	if err = tx.Commit().Error; err != nil {
		bt.err <- err
		return
	}
	err = RunOutsideSteps(e.UserInfo, &outsideMission)
	bt.err <- err
}

var payToOutSideMutex = new(sync.Mutex)

func (e *BtcChannelEvent) PayToOutsideOnChain(address string, amount float64) error {
	payToOutSideMutex.Lock()
	defer payToOutSideMutex.Unlock()

	if !verifyBtcAddress(address) {
		return fmt.Errorf("%s is not a valid btc address", address)
	}

	endAmount := amount + 2500

	limitType := custodyModels.LimitType{
		AssetId:      "00",
		TransferType: custodyModels.LimitTransferTypeOutside,
	}
	err := custodyLimit.CheckLimit(middleware.DB, e.UserInfo, &limitType, endAmount)
	if err != nil {
		return err
	}
	if !custodyBalance.CheckBtcBalance(middleware.DB, e.UserInfo, float64(endAmount)) {
		return NotSufficientFunds
	}
	serverBalance, err := rpc.GetBalance()
	if err != nil {
		return fmt.Errorf("get server balance error:%s", err.Error())
	}
	if serverBalance.AccountBalance["default"].ConfirmedBalance-10000 < int64(endAmount) {
		return fmt.Errorf("no enough server balance")
	}
	err = payBtcOnchain(e.UserInfo, address, amount)
	if err != nil {
		btlLog.CUST.Error(err.Error())
		return err
	}
	return nil
}

func (e *BtcChannelEvent) GetTransactionHistory(query *cBase.PaymentRequest) (*cBase.PaymentList, error) {

	if query.Page <= 0 {
		return nil, fmt.Errorf("page error")
	}

	db := middleware.DB
	var err error
	var a []models.Balance
	offset := (query.Page - 1) * query.PageSize
	q := db.Where("account_id = ? and asset_id = ?", e.UserInfo.Account.ID, "00")
	switch query.Away {
	case 0, 1:
		q = q.Where("away = ?", query.Away)
	default:
	}
	err = q.Order("created_at desc").
		Limit(query.PageSize).
		Offset(offset).
		Find(&a).Error
	if err != nil {
		btlLog.CUST.Error(err.Error())
		return nil, fmt.Errorf("query payment error")
	}
	var results cBase.PaymentList
	if len(a) > 0 {
		for i := range a {
			v := a[i]
			r := cBase.PaymentResponse{}
			r.Timestamp = v.CreatedAt.Unix()
			r.BillType = v.BillType
			r.Away = v.Away
			r.Target = v.Invoice
			if *v.Invoice == "award" && v.PaymentHash != nil {
				awardType := cBase.GetAwardType(*v.PaymentHash)
				r.Target = &awardType
			}
			if strings.HasPrefix(*v.Invoice, "ptn") {
				var ptn custodyPayTN.PayToNpubKey
				err := ptn.Decode(*v.Invoice)
				if err == nil {
					if v.Away == models.AWAY_IN {
						r.Target = &ptn.FromNpubKey
					} else {
						r.Target = &ptn.NpubKey
					}
				}
			}
			if r.BillType == models.BillTypePendingOder {
				if strings.HasPrefix(*v.PaymentHash, "stake") {
					var temp string
					if *v.Invoice == "pendingOderPay" {
						temp = "质押"
						r.Target = &temp
					} else if *v.Invoice == "pendingOderReceive" {
						temp = "赎回"
						r.Target = &temp
					}
				}
			}
			empty := ""
			r.Invoice = &empty
			r.Address = &empty
			if v.PaymentHash != nil {
				r.PaymentHash = v.PaymentHash
			} else {
				r.PaymentHash = &empty
			}
			r.Amount = v.Amount
			btcAssetId := "00"
			r.AssetId = &btcAssetId
			r.State = v.State
			r.Fee = uint64(v.ServerFee)
			results.PaymentList = append(results.PaymentList, r)
		}
	}
	return &results, nil
}
