package custodyAssets

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"strings"
	"time"
	"trade/btlLog"
	"trade/config"
	"trade/middleware"
	"trade/models"
	"trade/models/custodyModels"
	"trade/services/assetsyncinfo"
	"trade/services/btldb"
	caccount "trade/services/custodyAccount/account"
	cBase "trade/services/custodyAccount/custodyBase"
	"trade/services/custodyAccount/custodyBase/custodyFee"
	"trade/services/custodyAccount/custodyBase/custodyLimit"
	"trade/services/custodyAccount/custodyBase/custodyPayTN"
	"trade/services/custodyAccount/defaultAccount/custodyBtc"
	"trade/services/custodyAccount/defaultAccount/custodyBtc/mempool"
	rpc "trade/services/servicesrpc"
)

type AssetEvent struct {
	UserInfo *caccount.UserInfo
	AssetId  *string
}

func NewAssetEvent(UserName string, AssetId string) (*AssetEvent, error) {
	var (
		e   AssetEvent
		err error
	)
	e.UserInfo, err = caccount.GetUserInfo(UserName)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: %s", caccount.CustodyAccountGetErr, "userName不存在")
		}
		return nil, fmt.Errorf("%w: %w", caccount.CustodyAccountGetErr, err)
	}
	e.AssetId = &AssetId
	return &e, nil
}

func (e *AssetEvent) GetBalance() ([]cBase.Balance, error) {
	balance := GetAssetBalance(middleware.DB, e.UserInfo.Account.ID, *e.AssetId)
	balances := []cBase.Balance{
		{
			AssetId: *e.AssetId,
			Amount:  int64(balance),
		},
	}
	return balances, nil
}

func (e *AssetEvent) GetBalances() ([]cBase.Balance, error) {
	temp := GetAssetsBalances(middleware.DB, e.UserInfo.Account.ID)
	var balances []cBase.Balance
	for _, b := range *temp {
		balances = append(balances, cBase.Balance{
			AssetId: b.AssetId,
			Amount:  int64(b.Amount),
		})
	}
	return balances, nil
}

func (e *AssetEvent) GetCustodyAssetPermission(assetId, universe string) (*models.AssetSyncInfo, error) {
	r := assetsyncinfo.SyncInfoRequest{
		Id:       assetId,
		Universe: universe,
	}
	s, err := assetsyncinfo.GetAssetSyncInfo(&r)
	if err != nil {
		return nil, err
	}
	if s.AssetType == models.AssetTypeNFT {
		return nil, fmt.Errorf("NFT custody is not supported")
	}
	return s, nil
}

var CreateAddrErr = errors.New("CreateAddrErr")

func (e *AssetEvent) ApplyPayReq(Request cBase.PayReqApplyRequest) (cBase.PayReqApplyResponse, error) {
	var applyRequest *AssetAddressApplyRequest
	var ok bool
	if applyRequest, ok = Request.(*AssetAddressApplyRequest); !ok {
		return nil, errors.New("invalid apply request")
	}
	universe := config.GetConfig().ApiConfig.Tapd.UniverseHost
	//调用Lit节点发票申请接口
	addr, err := rpc.NewAddr(*e.AssetId, int(applyRequest.Amount), universe)
	if err != nil {
		btlLog.CUST.Error(err.Error())
		return nil, fmt.Errorf("%w: %s", CreateAddrErr, err.Error())
	}
	template := time.Now()
	expiry := 0
	//构建invoice对象
	var invoiceModel models.Invoice
	invoiceModel.UserID = e.UserInfo.User.ID
	invoiceModel.Invoice = addr.Encoded
	invoiceModel.AccountID = &e.UserInfo.Account.ID
	invoiceModel.AssetId = *e.AssetId
	invoiceModel.Amount = float64(addr.Amount)
	invoiceModel.Status = models.InvoiceStatusIsTaproot
	invoiceModel.CreateDate = &template
	invoiceModel.Expiry = &expiry
	//写入数据库
	err = btldb.CreateInvoice(&invoiceModel)
	if err != nil {
		btlLog.CUST.Error(err.Error(), models.ReadDbErr)
		return nil, models.ReadDbErr
	}
	return &AssetApplyAddress{
		Addr:   addr,
		Amount: applyRequest.Amount,
	}, nil
}

func (e *AssetEvent) SendPaymentToUser(receiverUserName string, amount float64, assetId string) error {
	//if !control.GetTransferControl("asset", control.TransferControlLocal) {
	//	return errors.New("当前服务调用失败，请稍后再试")
	//}
	//检查接收方是否存在
	var err error
	receiver, err := caccount.GetUserInfo(receiverUserName)
	if err != nil {
		btlLog.CUST.Warning("%s,UserName:%s", err.Error(), receiverUserName)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("%w: %s", caccount.CustodyAccountGetErr, "userName不存在")
		}
		return fmt.Errorf("%w: %w", caccount.CustodyAccountGetErr, err)
	}
	//检查余额是否足够
	limitType := custodyModels.LimitType{
		AssetId:      assetId,
		TransferType: custodyModels.LimitTransferTypeLocal,
	}
	err = custodyLimit.CheckLimit(middleware.DB, e.UserInfo, &limitType, amount)
	if err != nil {
		return err
	}

	//验证资产金额
	if !CheckAssetBalance(middleware.DB, e.UserInfo, assetId, amount) {
		return cBase.NotEnoughAssetFunds
	}
	if !custodyBtc.CheckBtcBalance(middleware.DB, e.UserInfo, float64(custodyFee.AssetInsideFee)) {
		return cBase.NotEnoughFeeFunds
	}
	//构建转账记录
	m := custodyModels.AccountInsideMission{
		AccountId:  e.UserInfo.Account.ID,
		AssetId:    assetId,
		Type:       custodyModels.AIMTypeAsset,
		ReceiverId: receiver.Account.ID,
		InvoiceId:  0,
		Amount:     amount,
		Fee:        float64(custodyFee.AssetInsideFee),
		FeeType:    custodyBtc.BtcId,
		State:      custodyModels.AIMStatePending,
	}
	custodyBtc.LogAIM(middleware.DB, &m)
	err = RunInsideStepByUserId(e.UserInfo, receiver, &m)
	if err != nil {
		return err
	}
	return nil
}

func (e *AssetEvent) SendPayment(payRequest cBase.PayPacket) error {
	var bt *AssetPacket
	var ok bool
	if bt, ok = payRequest.(*AssetPacket); !ok {
		return errors.New("invalid pay request")
	}
	bt.err = make(chan error, 1)
	//defer close(bt.err)

	err := bt.VerifyPayReq(e.UserInfo)
	if err != nil {
		return err
	}
	if bt.isInsideMission != nil {
		//if !control.GetTransferControl("asset", control.TransferControlLocal) {
		//	return errors.New("当前服务调用失败，请稍后再试")
		//}
		//发起本地转账
		bt.isInsideMission.err = bt.err
		go e.payToInside(bt)
	} else {
		//if !control.GetTransferControl("asset", control.TransferControlOnChain) {
		//	return errors.New("当前服务调用失败，请稍后再试")
		//}
		//发起外部转账
		go e.payToOutside(bt)
	}
	ctx, cancel := context.WithTimeout(context.Background(), cBase.Timeout)
	defer cancel()
	select {
	case <-ctx.Done():
		//超时处理
		go func(c chan error) {
			err := <-c
			if err != nil {
				btlLog.CUST.Error("btc sendPayment timeout:%s", err.Error())
			}
			close(c)
		}(bt.err)
		return cBase.TimeoutErr
	case err = <-bt.err:
		//错误处理
		return err
	}
}

func (e *AssetEvent) payToInside(bt *AssetPacket) {
	AssetId := hex.EncodeToString(bt.DecodePayReq.AssetId)
	m := custodyModels.AccountInsideMission{
		AccountId:  e.UserInfo.Account.ID,
		AssetId:    AssetId,
		Type:       custodyModels.AIMTypeAsset,
		ReceiverId: *bt.isInsideMission.insideInvoice.AccountID,
		InvoiceId:  bt.isInsideMission.insideInvoice.ID,
		Amount:     float64(bt.DecodePayReq.Amount),
		Fee:        float64(custodyFee.AssetInsideFee),
		FeeType:    custodyBtc.BtcId,
		State:      custodyModels.AIMStatePending,
	}
	custodyBtc.LogAIM(middleware.DB, &m)
	err := RunInsideStep(e.UserInfo, &m)
	bt.err <- err
}

func (e *AssetEvent) payToOutside(bt *AssetPacket) {
	tx, back := middleware.GetTx()
	defer back()
	var err error
	//获取资产id
	assetId := hex.EncodeToString(bt.DecodePayReq.AssetId)
	//更新额度限制
	limitType := custodyModels.LimitType{
		AssetId:      assetId,
		TransferType: custodyModels.LimitTransferTypeOutside,
	}
	err = custodyLimit.MinusLimit(tx, e.UserInfo, &limitType, float64(bt.DecodePayReq.Amount))
	if err != nil {
		bt.err <- fmt.Errorf("payToOutside limit error: %s", err.Error())
		return
	}
	//更新bills
	outsideBalance := models.Balance{
		AccountId: e.UserInfo.Account.ID,
		BillType:  models.BillTypeAssetTransfer,
		Away:      models.AWAY_OUT,
		Amount:    float64(bt.DecodePayReq.Amount),
		Unit:      models.UNIT_ASSET_NORMAL,
		ServerFee: uint64(mempool.GetCustodyAssetFee()),
		AssetId:   &assetId,
		Invoice:   &bt.PayReq,
		State:     models.STATE_UNKNOW,
		TypeExt: &models.BalanceTypeExt{
			Type: models.BTExtOnChannel,
		},
	}
	err = btldb.CreateBalance(tx, &outsideBalance)
	if err != nil {
		bt.err <- fmt.Errorf("payToOutside bill error: %s", err.Error())
		return
	}
	//收取费用
	_, err = LessAssetBalance(tx, e.UserInfo, outsideBalance.Amount, outsideBalance.ID, *outsideBalance.AssetId, custodyModels.ChangeTypeAssetPayOutside)
	if err != nil {
		bt.err <- fmt.Errorf("payToOutside asset balance error: %s", err.Error())
		return
	}
	err = custodyBtc.PayFee(tx, e.UserInfo, float64(mempool.GetCustodyAssetFee()), outsideBalance.ID)
	if err != nil {
		btlLog.CUST.Error("PayFee error:%s", err)
		return
	}
	// 创建对外转账记录
	outside := custodyModels.PayOutside{
		AccountID: e.UserInfo.Account.ID,
		AssetId:   assetId,
		Address:   bt.DecodePayReq.Encoded,
		Amount:    float64(bt.DecodePayReq.Amount),
		BalanceId: outsideBalance.ID,
		Status:    custodyModels.PayOutsideStatusPending,
	}
	err = btldb.CreatePayOutside(&outside)
	if err != nil {
		btlLog.CUST.Error("payToOutside db error")
	}
	if tx.Commit().Error != nil {
		btlLog.CUST.Error("payToOutside commit error")
		bt.err <- fmt.Errorf("payToOutside commit error")
		return
	}
	bt.err <- nil
	btlLog.CUST.Info("Create payToOutside mission success: id=%v,amount=%v", assetId, float64(bt.DecodePayReq.Amount))
}

func (e *AssetEvent) QueryPayReq() ([]*models.Invoice, error) {
	params := btldb.QueryParams{
		"UserID":  e.UserInfo.User.ID,
		"Status":  "10",
		"AssetId": *e.AssetId,
	}
	a, err := btldb.GenericQuery(&models.Invoice{}, params)
	if err != nil {
		btlLog.CUST.Error(err.Error())
		return nil, err
	}
	return a, nil
}

func (e *AssetEvent) QueryPayReqs() ([]*models.Invoice, error) {
	params := btldb.QueryParams{
		"UserID": e.UserInfo.User.ID,
		"Status": "10",
	}
	a, err := btldb.GenericQuery(&models.Invoice{}, params)
	if err != nil {
		btlLog.CUST.Error(err.Error())
		return nil, err
	}
	return a, nil
}

func (e *AssetEvent) GetTransactionHistory(query *cBase.PaymentRequest) (*cBase.PaymentList, error) {

	if query.Page <= 0 {
		return nil, fmt.Errorf("page error")
	}

	var a []models.Balance
	offset := (query.Page - 1) * query.PageSize
	q := middleware.DB.Where("account_id = ? AND asset_id = ?", e.UserInfo.Account.ID, query.AssetId)
	switch query.Away {
	case 0, 1:
		q = q.Where("away = ?", query.Away)
	default:
	}
	err := q.Order("created_at desc").
		Limit(query.PageSize).
		Offset(offset).
		Find(&a).Error
	if err != nil {
		btlLog.CUST.Error(err.Error())
		return nil, err
	}

	var results cBase.PaymentList
	if len(a) > 0 {
		for i := range a {
			if a[i].State == models.STATE_FAILED {
				continue
			}
			v := a[i]
			r := cBase.PaymentResponse{}
			r.Timestamp = v.CreatedAt.Unix()
			r.BillType = v.BillType
			r.Away = v.Away
			r.Target = v.Invoice
			r.PaymentHash = v.PaymentHash
			if *v.Invoice == "award" && v.PaymentHash != nil {
				awardType := cBase.GetAwardType(*v.PaymentHash)
				r.Target = &awardType
			}
			if strings.HasPrefix(*v.Invoice, "ptn") {
				var ptn custodyPayTN.PayToNpubKey
				err := ptn.Decode(*v.Invoice)
				if err == nil {
					r.Target = &ptn.NpubKey
				}
			}
			if r.BillType == models.BillTypePendingOder {
				if strings.HasPrefix(*v.PaymentHash, "stake/pay/") {
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
			r.Invoice = v.Invoice
			r.Address = v.Invoice
			r.Amount = v.Amount
			r.AssetId = a[i].AssetId
			r.State = v.State
			r.Fee = v.ServerFee
			results.PaymentList = append(results.PaymentList, r)
		}
	}
	return &results, nil
}
