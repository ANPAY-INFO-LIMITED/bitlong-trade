package lockPayment

import (
	"errors"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
	"trade/btlLog"
	"trade/middleware"
	"trade/models"
	cModels "trade/models/custodyModels"
	caccount "trade/services/custodyAccount/account"
	"trade/services/custodyAccount/defaultAccount/custodyAssets"
	"trade/services/custodyAccount/defaultAccount/custodyBtc"
)

var (
	ServerBusy    = errors.New("seriver is busy, please try again later")
	NoAwardType   = fmt.Errorf("no award type")
	AssetIdLock   = fmt.Errorf("award is lock")
	NoEnoughAward = fmt.Errorf("not enough award")
)

func PutInAwardLockBTC(usr *caccount.UserInfo, amount float64, memo *string, lockedId string) (*models.AccountAward, error) {
	mutex := GetLockPaymentMutex(usr.User.ID)
	mutex.Lock()
	defer mutex.Unlock()

	tx, back := middleware.GetTx()
	defer back()
	var err error

	// send btc award
	lockedBalance := cModels.LockBalance{}
	if err = tx.Where("account_id =? AND asset_id =?", usr.LockAccount.ID, btcId).First(&lockedBalance).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			tx.Rollback()
			btlLog.CUST.Error(err.Error())
			return nil, ServiceError
		}
		// Init Balance record
		lockedBalance.AssetId = btcId
		lockedBalance.AccountID = usr.LockAccount.ID
		lockedBalance.Amount = 0
	}
	lockedBalance.Amount += amount
	if err = tx.Save(&lockedBalance).Error; err != nil {
		tx.Rollback()
		btlLog.CUST.Error(err.Error())
		return nil, ServiceError
	}

	//create locked balance bill
	lockBill := cModels.LockBill{
		AccountID: usr.LockAccount.ID,
		AssetId:   btcId,
		Amount:    amount,
		LockId:    lockedId,
		BillType:  cModels.LockBillTypeAward,
	}
	if err = tx.Create(&lockBill).Error; err != nil {
		var mySQLErr *mysql.MySQLError
		if errors.As(err, &mySQLErr) {
			if mySQLErr.Number == 1062 {
				return nil, RepeatedLockId
			}
		}
		btlLog.CUST.Error(err.Error())
		return nil, ServiceError
	}

	// Build a database AccountAward
	award := models.AccountAward{
		AccountID: usr.Account.ID,
		AssetId:   btcId,
		Amount:    amount,
		Memo:      memo,
	}
	if err = tx.Create(&award).Error; err != nil {
		btlLog.CUST.Error(err.Error())
		return nil, err
	}

	//Build a database AwardIdempotent
	Idempotent := models.AccountAwardIdempotent{
		AwardId:    award.ID,
		Idempotent: lockedId,
	}
	if err = tx.Create(&Idempotent).Error; err != nil {
		var mySQLErr *mysql.MySQLError
		if errors.As(err, &mySQLErr) {
			if mySQLErr.Number == 1062 {
				return nil, RepeatedLockId
			}
		}
		btlLog.CUST.Error(err.Error())
		return nil, ServiceError
	}

	// Build a database  AccountAwardExt
	awardExt := models.AccountAwardExt{
		BalanceId:   lockBill.ID,
		AwardId:     award.ID,
		AccountType: models.LockedAccount,
	}
	if err = tx.Create(&awardExt).Error; err != nil {
		btlLog.CUST.Error(err.Error())
		return nil, err
	}

	//扣除admin账户对应的金额
	var adminUsr *caccount.UserInfo
	adminUsr, err = caccount.GetUserInfo("admin")
	if err != nil {
		btlLog.CUST.Error(err.Error())
		return nil, err
	}
	payAwardInvoice := "offerAward"
	payba := models.Balance{}
	payba.AccountId = adminUsr.Account.ID
	payba.Amount = amount
	payba.Unit = models.UNIT_SATOSHIS
	payba.BillType = models.BillTypeOfferAward
	payba.Away = models.AWAY_OUT
	payba.Invoice = &payAwardInvoice
	payba.PaymentHash = &lockedId
	payba.State = models.STATE_SUCCESS
	payba.TypeExt = &models.BalanceTypeExt{Type: models.BTExtOfferAward}
	err = tx.Create(&payba).Error
	if err != nil {
		btlLog.CUST.Error(err.Error())
		return nil, err
	}
	payAwardExt := models.AccountAwardExt{
		BalanceId:   payba.ID,
		AwardId:     award.ID,
		AccountType: models.LockedAccount,
	}
	if err = tx.Create(&payAwardExt).Error; err != nil {
		btlLog.CUST.Error(err.Error())
		return nil, err
	}
	_, err = custodyBtc.LessBtcBalance(tx, adminUsr, payba.Amount, payba.ID, cModels.ChangeTypeOfferAward)
	if err != nil {
		btlLog.CUST.Error(err.Error())
		return nil, err
	}

	tx.Commit()

	return &award, nil
}

func PutInAwardLockAsset(usr *caccount.UserInfo, assetId string, amount float64, memo *string, lockedId string) (*models.AccountAward, error) {
	mutex := GetLockPaymentMutex(usr.User.ID)
	mutex.Lock()
	defer mutex.Unlock()

	tx, back := middleware.GetTx()
	defer back()
	var err error

	// Check if the asset is award type
	var in models.AwardInventory
	err = tx.Where("asset_Id =? ", assetId).First(&in).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		btlLog.CUST.Error("err:%v", err)
		return nil, ServerBusy
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, NoAwardType
	}
	if in.Status != models.AwardInventoryAble {
		return nil, AssetIdLock
	}
	if in.Amount < amount {
		return nil, NoEnoughAward
	}

	// send btc award
	lockedBalance := cModels.LockBalance{}
	if err = tx.Where("account_id =? AND asset_id =?", usr.LockAccount.ID, assetId).First(&lockedBalance).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			tx.Rollback()
			btlLog.CUST.Error(err.Error())
			return nil, ServiceError
		}
		// Init Balance record
		lockedBalance.AssetId = assetId
		lockedBalance.AccountID = usr.LockAccount.ID
		lockedBalance.Amount = 0
	}
	lockedBalance.Amount += amount
	if err = tx.Save(&lockedBalance).Error; err != nil {
		tx.Rollback()
		btlLog.CUST.Error(err.Error())
		return nil, ServiceError
	}

	//create locked balance bill
	lockBill := cModels.LockBill{
		AccountID: usr.LockAccount.ID,
		AssetId:   assetId,
		Amount:    amount,
		LockId:    lockedId,
		BillType:  cModels.LockBillTypeAward,
	}
	if err = tx.Create(&lockBill).Error; err != nil {
		var mySQLErr *mysql.MySQLError
		if errors.As(err, &mySQLErr) {
			if mySQLErr.Number == 1062 {
				return nil, RepeatedLockId
			}
		}
		btlLog.CUST.Error(err.Error())
		return nil, ServiceError
	}

	// Build a database AccountAward
	award := models.AccountAward{
		AccountID: usr.Account.ID,
		AssetId:   assetId,
		Amount:    amount,
		Memo:      memo,
	}
	if err = tx.Create(&award).Error; err != nil {
		btlLog.CUST.Error(err.Error())
		return nil, err
	}
	//Build a database AwardIdempotent
	Idempotent := models.AccountAwardIdempotent{
		AwardId:    award.ID,
		Idempotent: lockedId,
	}
	if err = tx.Create(&Idempotent).Error; err != nil {
		var mySQLErr *mysql.MySQLError
		if errors.As(err, &mySQLErr) {
			if mySQLErr.Number == 1062 {
				return nil, RepeatedLockId
			}
		}
		btlLog.CUST.Error(err.Error())
		return nil, ServiceError
	}

	// Build a database  AccountAwardExt
	awardExt := models.AccountAwardExt{
		BalanceId:   lockBill.ID,
		AwardId:     award.ID,
		AccountType: models.LockedAccount,
	}
	if err = tx.Create(&awardExt).Error; err != nil {
		btlLog.CUST.Error(err.Error())
		return nil, err
	}
	//扣除admin账户对应的金额
	var adminUsr *caccount.UserInfo
	adminUsr, err = caccount.GetUserInfo("admin")
	if err != nil {
		btlLog.CUST.Error(err.Error())
		return nil, err
	}
	payAwardInvoice := "offerAward"
	payba := models.Balance{}
	payba.AccountId = adminUsr.Account.ID
	payba.Amount = amount
	payba.Unit = models.UNIT_ASSET_NORMAL
	payba.BillType = models.BillTypeOfferAward
	payba.Away = models.AWAY_OUT
	payba.AssetId = &assetId
	payba.Invoice = &payAwardInvoice
	payba.PaymentHash = &lockedId
	payba.State = models.STATE_SUCCESS
	payba.TypeExt = &models.BalanceTypeExt{Type: models.BTExtOfferAward}
	err = tx.Create(&payba).Error
	if err != nil {
		btlLog.CUST.Error(err.Error())
		return nil, err
	}
	payAwardExt := models.AccountAwardExt{
		BalanceId: payba.ID,
		AwardId:   award.ID,
	}
	if err = tx.Create(&payAwardExt).Error; err != nil {
		btlLog.CUST.Error(err.Error())
		return nil, err
	}
	_, err = custodyAssets.LessAssetBalance(tx, adminUsr, payba.Amount, payba.ID, assetId, cModels.ChangeTypeOfferAward)
	if err != nil {
		btlLog.CUST.Error(err.Error())
		return nil, err
	}
	tx.Commit()

	return &award, nil
}
