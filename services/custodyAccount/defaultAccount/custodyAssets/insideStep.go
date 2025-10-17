package custodyAssets

import (
	"errors"
	"fmt"
	"gorm.io/gorm"
	"math/rand"
	"time"
	"trade/btlLog"
	"trade/middleware"
	"trade/models"
	"trade/models/custodyModels"
	"trade/models/custodyModels/game"
	"trade/services/custodyAccount/account"
	"trade/services/custodyAccount/custodyBase/custodyLimit"
	"trade/services/custodyAccount/custodyBase/custodyPayTN"
	"trade/services/custodyAccount/defaultAccount/custodyBalance"
	"trade/services/custodyAccount/defaultAccount/custodyBtc"
	"trade/services/custodyAccount/defaultAccount/custodyBtc/mempool"
)

type invoiceInfo struct {
	Invoice string
	AssetId string
	Hash    *string
}

func RunInsideStep(usr *account.UserInfo, mission *custodyModels.AccountInsideMission) error {
	db := middleware.DB

	if usr == nil {
		var a models.Account
		if err := db.Where("id =?", mission.AccountId).First(&a).Error; err != nil {
			btlLog.CUST.Error("GetAccount error:%s", err)
			return err
		}
		usr, _ = account.GetUserInfo(a.UserName)
	}

	invoice := models.Invoice{}
	if err := db.Where("id =?", mission.InvoiceId).First(&invoice).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		btlLog.CUST.Error("GetInvoice error:%s", err)
		return err
	}
	i := invoiceInfo{
		Invoice: invoice.Invoice,
		AssetId: mission.AssetId,
		Hash:    nil,
	}

	for {
		InsideSteps(usr, mission, i)
		custodyBtc.LogAIM(middleware.DB, mission)
		switch {
		case mission.State == custodyModels.AIMStateSuccess:
			return nil
		case mission.State == custodyModels.AIMStateDone:
			return fmt.Errorf(mission.Error)
		case mission.Retries >= 30:
			return nil
		}
	}
}

func RunInsideStepByUserId(usr *account.UserInfo, receiveUsr *account.UserInfo, mission *custodyModels.AccountInsideMission) error {
	db := middleware.DB

	if usr == nil {
		var a models.Account
		if err := db.Where("id =?", mission.AccountId).First(&a).Error; err != nil {
			btlLog.CUST.Error("GetAccount error:%s", err)
			return err
		}
		usr, _ = account.GetUserInfo(a.UserName)
	}
	if receiveUsr == nil {
		var a models.Account
		if err := db.Where("id =?", mission.ReceiverId).First(&a).Error; err != nil {
			btlLog.CUST.Error("GetAccount error:%s", err)
			return err
		}
		receiveUsr, _ = account.GetUserInfo(a.UserName)
	}

	PTN := custodyPayTN.PayToNpubKey{
		NpubKey:     receiveUsr.User.Username,
		Amount:      mission.Amount,
		AssetId:     mission.AssetId,
		Time:        mission.CreatedAt.Unix(),
		FromNpubKey: usr.User.Username,
		Vision:      1,
	}
	invoice, _ := PTN.Encode()
	h, _ := custodyPayTN.HashEncodedString(invoice)
	i := invoiceInfo{
		Invoice: invoice,
		AssetId: mission.AssetId,
		Hash:    &h,
	}
	for {
		InsideSteps(usr, mission, i)
		custodyBtc.LogAIM(middleware.DB, mission)
		switch {
		case mission.State == custodyModels.AIMStateSuccess:
			return nil
		case mission.State == custodyModels.AIMStateDone:
			return fmt.Errorf(mission.Error)
		case mission.Retries >= 30:
			return nil
		}
	}
}

func InsideSteps(usr *account.UserInfo, mission *custodyModels.AccountInsideMission, i invoiceInfo) {
	var err error
	switch mission.State {
	case custodyModels.AIMStatePending:
		tx, back := middleware.GetTx()
		defer back()

		balance := getBillBalanceModel(usr, mission.Amount, i.AssetId, models.AWAY_OUT, i)
		if err = tx.Create(balance).Error; err != nil {
			btlLog.CUST.Error("CreateBillBalance error:%s", err)
			mission.Error = err.Error()
			mission.State = custodyModels.AIMStateDone
			return
		}
		mission.PayerBalanceId = balance.ID

		_, err = custodyBalance.LessAssetBalance(tx, usr, mission.Amount, mission.PayerBalanceId, i.AssetId, custodyModels.ChangeTypeAssetPayLocal)
		if err != nil {
			btlLog.CUST.Error("LessBtcBalance error:%s", err)
			mission.Error = err.Error()
			mission.State = custodyModels.AIMStateDone
			return
		}

		err = custodyBalance.PayFee(tx, usr, mission.Fee, mission.PayerBalanceId, &i.Invoice, i.Hash)
		if err != nil {
			btlLog.CUST.Error("PayFee error:%s", err)
			mission.Error = err.Error()
			mission.State = custodyModels.AIMStateDone
			return
		}
		balance.ServerFee = mission.Fee

		err = tx.Save(balance).Error
		if err != nil {
			btlLog.CUST.Error("SaveBillBalance error:%s", err)
			mission.Error = err.Error()
			mission.State = custodyModels.AIMStateDone
			return
		}
		mission.State = custodyModels.AIMStatePaid
		tx.Commit()

		go func() {

			limitType := custodyModels.LimitType{
				AssetId:      i.AssetId,
				TransferType: custodyModels.LimitTransferTypeLocal,
			}
			err = custodyLimit.MinusLimit(middleware.DB, usr, &limitType, mission.Amount+mission.Fee)
			if err != nil {
				btlLog.CUST.Error("额度限制未正常更新:%s", err.Error())
				btlLog.CUST.Error("error PayInsideId:%v", mission.ID)
			}
		}()
		return

	case custodyModels.AIMStatePaid:

		var a models.Account
		if err = middleware.DB.Where("id =?", mission.ReceiverId).First(&a).Error; err != nil {
			btlLog.CUST.Error("GetAccount error:%s", err)
			mission.Retries += 1
			mission.Error = err.Error()
			return
		}
		rusr, err := account.GetUserInfo(a.UserName)
		if err != nil {
			btlLog.CUST.Error("GetUserInfo error:%s", err)
			mission.Retries += 1
			mission.Error = err.Error()
			return
		}

		tx, back := middleware.GetTx()
		defer back()
		rBalance := getBillBalanceModel(rusr, mission.Amount, i.AssetId, models.AWAY_IN, i)
		if err = tx.Create(rBalance).Error; err != nil {
			btlLog.CUST.Error("CreateBillBalance error:%s", err)
			mission.Retries += 1
			mission.Error = err.Error()
			return
		}
		mission.ReceiverBalanceId = rBalance.ID

		_, err = custodyBalance.AddAssetBalance(tx, rusr, mission.Amount, rBalance.ID, i.AssetId, custodyModels.ChangeTypeAssetReceiveLocal)
		if err != nil {
			mission.Retries += 1
			mission.Error = err.Error()
			return
		}
		mission.State = custodyModels.AIMStateSuccess
		if rusr.Account.Type == models.GameReceiveAccount {
			if i.AssetId == mempool.GameAssetId && rusr.Account.UserName == mempool.GameUser {
				gameRecharge := game.Recharge{}
				nowTime := time.Now()
				rand.Seed(time.Now().UnixNano())
				randomNumber := rand.Intn(9000) + 1000
				gameRecharge.OrderId = fmt.Sprintf("41%d%d", nowTime.Unix(), randomNumber)
				gameRecharge.Npubkey = usr.Account.UserName
				gameRecharge.Amount = mission.Amount
				gameRecharge.AssetId = i.AssetId
				gameRecharge.GameAccountId = rusr.Account.ID
				gameRecharge.BalanceId = mission.ReceiverBalanceId
				err = tx.Create(&gameRecharge).Error
				if err != nil {
					btlLog.CUST.Error("CreateGameRecharge error:%s", err)
					mission.Error = err.Error()
					return
				}
				defer func() {
					mempool.PushTxRecord(&gameRecharge)
				}()
			}
		}
		tx.Commit()

		flag := i.Invoice[0:2]
		if flag == "ln" {
			go func() {
				err = middleware.DB.Model(&models.Invoice{}).
					Where("id =?", mission.InvoiceId).
					Updates(&models.Invoice{Status: models.InvoiceStatusLocal}).Error
			}()
		}
		return
	}
}

func getBillBalanceModel(usr *account.UserInfo, amount float64, assetId string, away models.BalanceAway, invoice invoiceInfo) *models.Balance {
	ba := models.Balance{}
	ba.AccountId = usr.Account.ID
	ba.Amount = amount
	ba.AssetId = &assetId
	ba.Unit = models.UNIT_ASSET_NORMAL
	ba.BillType = models.BillTypeAssetTransfer
	ba.Away = away
	ba.Invoice = &invoice.Invoice
	ba.PaymentHash = invoice.Hash
	ba.State = models.STATE_SUCCESS
	ba.TypeExt = &models.BalanceTypeExt{Type: models.BTExtLocal}
	return &ba
}

func LoadAIMMission() {
	var missions []custodyModels.AccountInsideMission
	middleware.DB.Where("type = 'asset' AND (state =? OR state =?)", custodyModels.AIMStatePending, custodyModels.AIMStatePaid).Find(&missions)
	for _, m := range missions {
		_ = RunInsideStep(nil, &m)
	}
}
