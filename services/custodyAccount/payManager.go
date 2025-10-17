package custodyAccount

import (
	"context"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"log"
	"math"
	"time"
	"trade/btlLog"
	"trade/config"
	"trade/middleware"
	"trade/models"
	"trade/models/custodyModels"
	"trade/services/btldb"
	"trade/services/custodyAccount/account"
	cBase "trade/services/custodyAccount/custodyBase"
	"trade/services/custodyAccount/defaultAccount/costodyRecive"
	"trade/services/custodyAccount/defaultAccount/custodyAssets"
	"trade/services/custodyAccount/defaultAccount/custodyBalance"
	"trade/services/custodyAccount/defaultAccount/custodyBtc"
	"trade/services/custodyAccount/defaultAccount/custodyGame"
	"trade/services/custodyAccount/lockPayment"
)

var (
	AdminUserInfo *account.UserInfo
)

type ApplyRequest struct {
	Amount int64  `json:"amount"`
	Memo   string `json:"memo"`
}

type PayInvoiceRequest struct {
	Invoice string `json:"invoice"`
}

type PaymentRequest struct {
	AssetId  string `json:"asset_id"`
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
}

type DecodeInvoiceRequest struct {
	Invoice string `json:"invoice"`
}

func CustodyStart(ctx context.Context, cfg *config.Config) bool {
	timestart := time.Now()

	if !checkAdminAccount() {
		btlLog.CUST.Error("Admin account is not set")
		return false
	}
	{

		costodyRecive.InvoiceServer.Start(ctx)

		custodyBtc.LoadAOMMission()
		custodyBtc.LoadAIMMission()

		custodyBtc.BtcRechargeOnChainDaemon()
	}
	timeend := time.Now()
	log.Printf("btcServer time:%v\n", timeend.Sub(timestart))

	{

		costodyRecive.AddressServer.Start(ctx)

		custodyAssets.GoOutsideMission()

		custodyAssets.LoadAIMMission()
		custodyAssets.LoadAOMAssetMission()
	}
	timeend = time.Now()
	log.Printf("assetServer time:%v\n", timeend.Sub(timestart))
	{

		custodyGame.StartGamePushTxRecordDaemon()
	}
	if cfg.CustodyConfig.ClearBlockAccountBalance {

	}
	timeend = time.Now()
	log.Printf("CustodyStart time:%v\n", timeend.Sub(timestart))

	return true
}

func checkAdminAccount() bool {
	adminUser, err := btldb.ReadUserByUsername("admin")
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			btlLog.CUST.Error("CheckAdminAccount failed:%s", err)
			return false
		}

		adminUser.Username = "admin"
		adminUser.Password = "admin"
		err = btldb.CreateUser(adminUser)
		if err != nil {
			btlLog.CUST.Error("create AdminUser failed:%s", err)
			return false
		}
	}

	adminAccount, err := account.GetUserInfo("admin")
	if err != nil {
		btlLog.CUST.Error("CheckAdminAccount failed:%s", err)
		return false
	}
	AdminUserInfo = adminAccount

	btlLog.CUST.Info("admin user id:%d", AdminUserInfo.User.ID)
	btlLog.CUST.Info("admin account id:%d", AdminUserInfo.Account.ID)
	btlLog.CUST.Info("admin lockAccount id:%d", AdminUserInfo.LockAccount.ID)
	return true
}

func LockPaymentToPaymentList(usr *account.UserInfo, assetId string, pageNum, pageSize, away int) (*cBase.PaymentList, error) {
	btc, err := lockPayment.ListTransferBTC(usr, assetId, pageNum, pageSize, away)
	if err != nil {
		return nil, err
	}
	db := middleware.DB
	var list cBase.PaymentList
	for i := range btc {
		v := btc[i]
		r := cBase.PaymentResponse{}
		r.Timestamp = v.CreatedAt.Unix()

		switch v.BillType {
		case custodyModels.LockBillTypeLock:
			r.Away = models.AWAY_IN
		case custodyModels.LockBillTypeAward:
			r.Away = models.AWAY_IN
			var awardExt models.AccountAwardExt
			db.Where("balance_id =? and account_type =1", v.ID).First(&awardExt)
			var award models.AccountAward
			db.Where("id =?", awardExt.AwardId).First(&award)
			v.LockId = cBase.GetAwardType(*award.Memo)

		default:
			r.Away = models.AWAY_OUT
		}
		r.BillType = models.LockedTransfer
		empty := ""
		r.Invoice = &empty
		r.Address = &empty
		r.Target = &v.LockId
		r.PaymentHash = &v.LockId
		r.Amount = v.Amount
		r.AssetId = &v.AssetId
		r.State = models.STATE_SUCCESS
		r.Fee = 0
		list.PaymentList = append(list.PaymentList, r)
	}
	return &list, nil
}

func ClearBtcToAsset() {
	db := middleware.DB
	var btcs []custodyModels.AccountBtcBalance
	err := db.Where("Amount > 0").Find(&btcs).Error
	if err != nil {
		btlLog.CUST.Error("ClearBtcToAsset failed:%s", err)
		return
	}
	adminacc, err = account.GetUserInfo("admin")
	if err != nil {
		btlLog.CUST.Error("ClearBtcToAsset failed:%s", err)
		return
	}

	for _, v := range btcs {
		for account.GetUserNum() > 1500 {
			time.Sleep(time.Second * 30)
		}
		if v.AccountId == 1 || v.AccountId == 2 || v.AccountId == 188740 || v.AccountId == 178821 {
			continue
		}
		acc := models.Account{}
		err := db.Where("id =?", v.AccountId).First(&acc).Error
		if err != nil {
			btlLog.CUST.Error("ClearBtcToAsset failed:%s", err)
			continue
		}
		userinfo, err := account.GetUserInfo(acc.UserName)
		if err != nil {
			btlLog.CUST.Error("ClearBtcToAsset failed:%s", err)
			continue
		}
		runReplace(userinfo, v.Amount)
	}
}

var adminacc *account.UserInfo

func runReplace(userinfo *account.UserInfo, amount float64) {
	tx, back := middleware.GetTx()
	defer back()
	assetId := "47ed120d4b173eb79ba46cd1959bb9c881cb69332cf8a21336110bda05402308"

	Invoice := "ReplaceAssetToPhinx"
	btcHash := fmt.Sprintf("ReplaceAssetToPhinx/btc/%v", userinfo.Account.ID)
	BtcBalance := models.Balance{
		AccountId:   userinfo.Account.ID,
		BillType:    models.BillTypeReplaceAsset,
		Away:        models.AWAY_OUT,
		Amount:      amount,
		Unit:        models.UNIT_SATOSHIS,
		ServerFee:   0,
		Invoice:     &Invoice,
		PaymentHash: &btcHash,
		State:       models.STATE_SUCCESS,
		TypeExt: &models.BalanceTypeExt{
			Type: models.BTEReplace,
		},
	}
	if err := tx.Create(&BtcBalance).Error; err != nil {
		btlLog.CUST.Error("runReplace failed:%s", err)
		return
	}
	_, err := custodyBalance.LessBtcBalance(tx, userinfo, amount, BtcBalance.ID, custodyModels.ChangeTypeReplaceAsset)
	if err != nil {
		btlLog.CUST.Error("runReplace failed:%s", err)
		return
	}
	adminBtcbalance := models.Balance{
		AccountId:   adminacc.Account.ID,
		BillType:    models.BillTypeReplaceAsset,
		Away:        models.AWAY_IN,
		Amount:      amount,
		Unit:        models.UNIT_SATOSHIS,
		ServerFee:   0,
		Invoice:     &Invoice,
		PaymentHash: &btcHash,
		State:       models.STATE_SUCCESS,
		TypeExt: &models.BalanceTypeExt{
			Type: models.BTEReplace,
		},
	}
	if err := tx.Create(&adminBtcbalance).Error; err != nil {
		btlLog.CUST.Error("runReplace failed:%s", err)
		return
	}
	_, err = custodyBalance.AddBtcBalance(tx, adminacc, amount, adminBtcbalance.ID, custodyModels.ChangeTypeReplaceAsset)
	if err != nil {
		btlLog.CUST.Error("runReplace failed:%s", err)
		return
	}

	assetHash := fmt.Sprintf("ReplaceAssetToPhinx/asset/%v", userinfo.Account.ID)
	assetAmount := math.Floor(amount / 2)
	AssetBalance := models.Balance{
		AccountId:   userinfo.Account.ID,
		BillType:    models.BillTypeReplaceAsset,
		Away:        models.AWAY_IN,
		Amount:      assetAmount,
		Unit:        models.UNIT_ASSET_NORMAL,
		ServerFee:   0,
		AssetId:     &assetId,
		Invoice:     &Invoice,
		PaymentHash: &assetHash,
		State:       models.STATE_SUCCESS,
		TypeExt: &models.BalanceTypeExt{
			Type: models.BTEReplace,
		},
	}
	if err := tx.Create(&AssetBalance).Error; err != nil {
		btlLog.CUST.Error("runReplace failed:%s", err)
		return
	}
	_, err = custodyBalance.AddAssetBalance(tx, userinfo, assetAmount, AssetBalance.ID, assetId, custodyModels.ChangeTypeReplaceAsset)
	if err != nil {
		btlLog.CUST.Error("runReplace failed:%s", err)
		return
	}

	adminAssetBalance := models.Balance{
		AccountId:   adminacc.Account.ID,
		BillType:    models.BillTypeReplaceAsset,
		Away:        models.AWAY_OUT,
		Amount:      assetAmount,
		Unit:        models.UNIT_ASSET_NORMAL,
		ServerFee:   0,
		AssetId:     &assetId,
		Invoice:     &Invoice,
		PaymentHash: &assetHash,
		State:       models.STATE_SUCCESS,
		TypeExt: &models.BalanceTypeExt{
			Type: models.BTEReplace,
		},
	}
	if err := tx.Create(&adminAssetBalance).Error; err != nil {
		btlLog.CUST.Error("runReplace failed:%s", err)
		return
	}
	_, err = custodyBalance.LessAssetBalance(tx, adminacc, assetAmount, adminAssetBalance.ID, assetId, custodyModels.ChangeTypeReplaceAsset)
	if err != nil {
		btlLog.CUST.Error("runReplace failed:%s", err)
		return
	}
	tx.Commit()
}
