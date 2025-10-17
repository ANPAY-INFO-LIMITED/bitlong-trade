package custodyAssets

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
	"trade/btlLog"
	"trade/config"
	"trade/middleware"
	"trade/models"
	"trade/models/custodyModels"
	"trade/models/custodyModels/custodyswap"
	"trade/services/assetsyncinfo"
	"trade/services/btldb"
	caccount "trade/services/custodyAccount/account"
	cBase "trade/services/custodyAccount/custodyBase"
	"trade/services/custodyAccount/custodyBase/control"
	"trade/services/custodyAccount/custodyBase/custodyFee"
	"trade/services/custodyAccount/custodyBase/custodyLimit"
	"trade/services/custodyAccount/custodyBase/custodyPayTN"
	"trade/services/custodyAccount/defaultAccount/custodyBalance"
	"trade/services/custodyAccount/defaultAccount/custodyBtc"
	"trade/services/custodyAccount/defaultAccount/custodyBtc/mempool"
	"trade/services/custodyAccount/defaultAccount/swap"
	rpc "trade/services/servicesrpc"

	"github.com/lightninglabs/taproot-assets/rfqmsg"
	"gorm.io/gorm"
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
	balance := custodyBalance.GetAssetBalance(middleware.DB, e.UserInfo.Account.ID, *e.AssetId)
	balances := []cBase.Balance{
		{
			AssetId: *e.AssetId,
			Amount:  int64(balance),
		},
	}
	return balances, nil
}

func (e *AssetEvent) GetBalances() ([]cBase.Balance, error) {
	temp, err := custodyBalance.GetAssetsBalances(middleware.DB, e.UserInfo.Account.ID)
	if err != nil {
		return nil, err
	}
	var balances []cBase.Balance
	if temp != nil && len(*temp) > 0 {
		for _, b := range *temp {
			balances = append(balances, cBase.Balance{
				AssetId: b.AssetId,
				Amount:  int64(b.Amount),
			})
		}
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
var CreateInvoiceErr = errors.New("CreateInvoiceErr")

func (e *AssetEvent) ApplyPayReq(applyRequest *AssetAddressApplyRequest) (*AssetApplyAddress, error) {
	universe := config.GetConfig().ApiConfig.Tapd.UniverseHost

	addr, err := rpc.NewAddr(*e.AssetId, int(applyRequest.Amount), universe)
	if err != nil {
		btlLog.CUST.Error(err.Error())
		return nil, fmt.Errorf("%w: %s", CreateAddrErr, err.Error())
	}
	template := time.Now()
	expiry := 0

	var invoiceModel models.Invoice
	invoiceModel.UserID = e.UserInfo.User.ID
	invoiceModel.Invoice = addr.Encoded
	invoiceModel.AccountID = &e.UserInfo.Account.ID
	invoiceModel.AssetId = *e.AssetId
	invoiceModel.Amount = float64(addr.Amount)
	invoiceModel.Status = models.InvoiceStatusIsTaproot
	invoiceModel.CreateDate = &template
	invoiceModel.Expiry = &expiry

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

func (e *AssetEvent) ApplyChannelPayReq(applyRequest *AssetInvoiceApplyResponse) (*AssetApplyInvoice, error) {
	PeerKeys, err := GetUsableChannelPeer(*e.AssetId, 0, uint64(applyRequest.Amount))
	if err != nil {
		return nil, err
	}

	invoice, err := rpc.AddAssetChannelInvoice(*e.AssetId, uint64(applyRequest.Amount), PeerKeys[0], "")
	if err != nil {
		btlLog.CUST.Error(err.Error())
		return nil, fmt.Errorf("%w: %s", CreateInvoiceErr, err.Error())
	}
	tx, back := middleware.GetTx()
	defer back()

	template := time.Now()
	expiry := int(int64(invoice.AcceptedBuyQuote.Expiry) - template.Unix())
	var invoiceModel models.Invoice
	invoiceModel.UserID = e.UserInfo.User.ID
	invoiceModel.Invoice = invoice.InvoiceResult.GetPaymentRequest()
	invoiceModel.AccountID = &e.UserInfo.Account.ID
	invoiceModel.AssetId = *e.AssetId
	invoiceModel.Amount = float64(applyRequest.Amount)
	invoiceModel.Status = models.InvoiceStatusPending
	invoiceModel.CreateDate = &template
	invoiceModel.Expiry = &expiry
	err = tx.Create(&invoiceModel).Error
	if err != nil {
		btlLog.CUST.Error(err.Error(), models.ReadDbErr)
		return nil, models.ReadDbErr
	}

	rfqInfoStr := invoice.AcceptedBuyQuote.String()
	rfqInfo := models.InvoiceRfqInfo{
		InvoiceId: invoiceModel.ID,
		RfqInfo:   rfqInfoStr,
	}
	err = tx.Create(&rfqInfo).Error
	if err != nil {
		btlLog.CUST.Error(err.Error(), models.ReadDbErr)
		return nil, models.ReadDbErr
	}
	tx.Commit()

	return &AssetApplyInvoice{RfqInfo: invoice.AcceptedBuyQuote,
		Invoice: invoice.InvoiceResult,
		Amount:  applyRequest.Amount}, nil
}

func (e *AssetEvent) SendPaymentToUser(receiverUserName string, amount float64, assetId string) error {
	if !control.GetTransferControl("asset", control.TransferControlLocal) {
		return errors.New("当前服务调用失败，请稍后再试")
	}

	var err error
	limitType := custodyModels.LimitType{
		AssetId:      assetId,
		TransferType: custodyModels.LimitTransferTypeLocal,
	}
	err = custodyLimit.CheckLimit(middleware.DB, e.UserInfo, &limitType, amount)
	if err != nil {
		return err
	}

	if !custodyBalance.CheckAssetBalance(middleware.DB, e.UserInfo, assetId, amount) {
		return cBase.NotEnoughAssetFunds
	}

	receiver, err := caccount.GetUserInfo(receiverUserName)
	if err != nil {
		btlLog.CUST.Warning("%s,UserName:%s", err.Error(), receiverUserName)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("%w: %s", caccount.CustodyAccountGetErr, "userName不存在")
		}
		return fmt.Errorf("%w: %w", caccount.CustodyAccountGetErr, err)
	}

	reciveList, character, err := swap.GetReceiveAssetId(receiverUserName, assetId)
	if err != nil {
		return fmt.Errorf("获取对方可接收资产列表失败")
	}
	assetIdSupplier, _ := swap.GetSupplier(assetId)
	if assetIdSupplier != "" && assetIdSupplier == receiverUserName || character == custodyswap.ReceiveCharacterConsumer {
		reciveList.AssetList = []string{assetId}
	}
	switch {
	case len(reciveList.AssetList) == 0:
		return fmt.Errorf("对方不接收任何资产")
	case len(reciveList.AssetList) == 1 && reciveList.AssetList[0] == assetId:

		if !custodyBalance.CheckBtcBalance(middleware.DB, e.UserInfo, float64(custodyFee.AssetInsideFee)) {
			return cBase.NotEnoughFeeFunds
		}

		m := custodyModels.AccountInsideMission{
			AccountId:  e.UserInfo.Account.ID,
			AssetId:    assetId,
			Type:       custodyModels.AIMTypeAsset,
			ReceiverId: receiver.Account.ID,
			InvoiceId:  0,
			Amount:     amount,
			FeeType:    custodyBalance.BtcId,
			State:      custodyModels.AIMStatePending,
		}
		assetFee, err := custodyBalance.GetAssetFee(assetId)
		if err != nil {
			return fmt.Errorf("GetAssetFee error: %s", err.Error())
		}
		m.Fee = assetFee
		custodyBtc.LogAIM(middleware.DB, &m)
		err = RunInsideStepByUserId(e.UserInfo, receiver, &m)
		if err != nil {
			return err
		}
		return nil
	case len(reciveList.AssetList) > 1,
		len(reciveList.AssetList) == 1 && reciveList.AssetList[0] != assetId:
		err := swap.PayBySwap(&swap.PTNSQuest{
			Payer:             e.UserInfo,
			PayAsset:          assetId,
			PayAmount:         amount,
			Receiver:          receiver,
			ReceiverAssetList: reciveList,
		})
		if err != nil {
			return fmt.Errorf("pay failed:%s", err)
		}
		return nil
	}
	return fmt.Errorf("未知错误:reciveList:%v，%v", reciveList, err)
}

func (e *AssetEvent) SendPayment(bt *AssetPacket) error {
	bt.err = make(chan error, 1)

	err := bt.VerifyPayReq(e)
	if err != nil {
		return err
	}
	if bt.isInsideMission != nil {
		if !control.GetTransferControl("asset", control.TransferControlLocal) {
			return errors.New("当前服务调用失败，请稍后再试")
		}

		bt.isInsideMission.err = bt.err
		go e.payToInside(bt)
	} else {

		if !control.GetTransferControl("asset", control.TransferControlOnChain) {
			return errors.New("当前服务调用失败，请稍后再试")
		}

		switch {
		case bt.DecodeAddr != nil:

			go e.payToOutsideOnChain(bt)
		case bt.DecodeInvoice != nil:

			go e.payToOutsideInChannel(bt)
		default:
			return errors.New("decodeAddr and DecodeInvoice is nil,invoice is invalid")
		}
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
	case err = <-bt.err:

		return err
	}
}

func (e *AssetEvent) payToInside(bt *AssetPacket) {
	m := custodyModels.AccountInsideMission{
		AccountId:  e.UserInfo.Account.ID,
		Type:       custodyModels.AIMTypeAsset,
		ReceiverId: *bt.isInsideMission.insideInvoice.AccountID,
		AssetId:    bt.isInsideMission.insideInvoice.AssetId,
		InvoiceId:  bt.isInsideMission.insideInvoice.ID,
		FeeType:    custodyBalance.BtcId,
		Amount:     bt.isInsideMission.insideInvoice.Amount,
		State:      custodyModels.AIMStatePending,
	}
	assetFee, err := custodyBalance.GetAssetFee(m.AssetId)
	if err != nil {
		bt.err <- fmt.Errorf("GetAssetFee error: %s", err.Error())
		return
	}
	m.Fee = assetFee

	custodyBtc.LogAIM(middleware.DB, &m)
	err = RunInsideStep(e.UserInfo, &m)
	bt.err <- err
}

func (e *AssetEvent) payToOutsideOnChain(bt *AssetPacket) {
	tx, back := middleware.GetTx()
	defer back()
	var err error

	assetId := hex.EncodeToString(bt.DecodeAddr.AssetId)

	limitType := custodyModels.LimitType{
		AssetId:      assetId,
		TransferType: custodyModels.LimitTransferTypeOutside,
	}
	err = custodyLimit.MinusLimit(tx, e.UserInfo, &limitType, float64(bt.DecodeAddr.Amount))
	if err != nil {
		bt.err <- fmt.Errorf("payToOutsideOnChain limit error: %s", err.Error())
		return
	}

	outsideBalance := models.Balance{
		AccountId: e.UserInfo.Account.ID,
		BillType:  models.BillTypeAssetTransfer,
		Away:      models.AWAY_OUT,
		Amount:    float64(bt.DecodeAddr.Amount),
		Unit:      models.UNIT_ASSET_NORMAL,
		ServerFee: float64(mempool.GetCustodyAssetFee()),
		AssetId:   &assetId,
		Invoice:   &bt.PayReq,
		State:     models.STATE_UNKNOW,
		TypeExt: &models.BalanceTypeExt{
			Type: models.BTExtOnChannel,
		},
	}
	err = btldb.CreateBalance(tx, &outsideBalance)
	if err != nil {
		bt.err <- fmt.Errorf("payToOutsideOnChain bill error: %s", err.Error())
		return
	}

	_, err = custodyBalance.LessAssetBalance(tx, e.UserInfo, outsideBalance.Amount, outsideBalance.ID, *outsideBalance.AssetId, custodyModels.ChangeTypeAssetPayOutside)
	if err != nil {
		bt.err <- fmt.Errorf("payToOutsideOnChain asset balance error: %s", err.Error())
		return
	}
	err = custodyBalance.PayFee(tx, e.UserInfo, float64(outsideBalance.ServerFee), outsideBalance.ID, &bt.PayReq, nil)
	if err != nil {
		btlLog.CUST.Error("PayFee error:%s", err)
		return
	}

	outside := custodyModels.PayOutside{
		AccountID: e.UserInfo.Account.ID,
		AssetId:   assetId,
		Address:   bt.DecodeAddr.Encoded,
		Amount:    float64(bt.DecodeAddr.Amount),
		BalanceId: outsideBalance.ID,
		Status:    custodyModels.PayOutsideStatusPending,
	}
	err = btldb.CreatePayOutside(&outside)
	if err != nil {
		btlLog.CUST.Error("payToOutsideOnChain db error")
	}
	if tx.Commit().Error != nil {
		btlLog.CUST.Error("payToOutsideOnChain commit error")
		bt.err <- fmt.Errorf("payToOutsideOnChain commit error")
		return
	}
	bt.err <- nil
	btlLog.CUST.Info("Create payToOutsideOnChain mission success: id=%v,amount=%v", assetId, float64(bt.DecodeAddr.Amount))
}

func (e *AssetEvent) payToOutsideInChannel(bt *AssetPacket) {
	tx, back := middleware.GetTx()
	defer back()

	var balanceModel models.Balance
	balanceModel.AccountId = e.UserInfo.Account.ID
	balanceModel.BillType = models.BillTypeAssetTransfer
	balanceModel.Away = models.AWAY_OUT
	balanceModel.Amount = float64(bt.DecodeInvoice.AssetAmount)
	balanceModel.Unit = models.UNIT_ASSET_NORMAL
	balanceModel.AssetId = e.AssetId
	balanceModel.Invoice = &bt.PayReq
	balanceModel.PaymentHash = &bt.DecodeInvoice.PayReq.PaymentHash
	balanceModel.State = models.STATE_UNKNOW
	balanceModel.TypeExt = &models.BalanceTypeExt{Type: models.BTExtOnChannel}
	err := btldb.CreateBalance(tx, &balanceModel)
	if err != nil {
		btlLog.CUST.Error(err.Error())
	}

	outsideMission := custodyModels.AccountOutsideMission{
		AccountId: e.UserInfo.Account.ID,
		AssetId:   *e.AssetId,
		Type:      custodyModels.AOMTypeAsset,
		Target:    bt.PayReq,
		Hash:      bt.DecodeInvoice.PayReq.PaymentHash,
		Amount:    float64(bt.DecodeInvoice.AssetAmount),
		FeeLimit:  float64(bt.DecodeInvoice.PayReq.NumSatoshis / 2),
		BalanceId: balanceModel.ID,
		State:     custodyModels.AOMStatePending,
	}
	tx.Save(&outsideMission)
	if err = tx.Commit().Error; err != nil {
		bt.err <- err
		return
	}
	err = runOutsideSteps(e.UserInfo, &outsideMission)
	bt.err <- err
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

func (e *AssetEvent) QueryChannelPayReq() ([]*models.Invoice, error) {
	db := middleware.DB
	invoices := make([]*models.Invoice, 0)
	err := db.Where("user_id = ? AND asset_id = ? AND status <> 10", e.UserInfo.User.ID, *e.AssetId).Find(&invoices).Error
	if err != nil {
		btlLog.CUST.Error(err.Error())
		return nil, err
	}
	return invoices, nil
}

func (e *AssetEvent) QueryChannelPayReqs() ([]*models.Invoice, error) {
	db := middleware.DB
	invoices := make([]*models.Invoice, 0)
	err := db.Where("user_id = ? AND asset_id <> '00' AND status <> 10", e.UserInfo.User.ID).Find(&invoices).Error
	if err != nil {
		btlLog.CUST.Error(err.Error())
		return nil, err
	}
	return invoices, nil
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
			r.AssetId = a[i].AssetId
			r.State = v.State
			r.Fee = uint64(v.ServerFee)
			results.PaymentList = append(results.PaymentList, r)
		}
	}
	return &results, nil
}

func GetUsableChannelPeer(assetId string, away int, amount uint64) ([]string, error) {
	resp, err := rpc.ListChannels(true, true)
	if err != nil {
		return nil, err
	}
	var peers []string
	for _, channel := range resp.GetChannels() {
		if channel.CustomChannelData != nil {
			var customData rfqmsg.JsonAssetChannel
			if err := json.Unmarshal(channel.CustomChannelData, &customData); err != nil {
				btlLog.CUST.Error("\n chanlist unmarshal \n %v", err)
				continue
			}
			switch away {
			case 0:
				for _, asset := range customData.RemoteAssets {
					if asset.AssetID == assetId && asset.Amount > amount && channel.RemoteBalance > 1600 {
						peers = append(peers, channel.RemotePubkey)
					}
				}
			case 1:
				for _, asset := range customData.LocalAssets {
					if asset.AssetID == assetId && asset.Amount > amount && channel.LocalBalance > 1600 {
						peers = append(peers, channel.RemotePubkey)
					}
				}
			}
		}
	}
	if len(peers) == 0 {
		return nil, fmt.Errorf("not found UsableChannelPeer")
	}
	return peers, nil
}
