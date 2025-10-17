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
	"trade/services/custodyAccount/defaultAccount/custodyBalance"
)

func GetAssetBalance(usr *caccount.UserInfo, assetId string) (err error, unlock float64, locked float64, tag1 float64) {
	db := middleware.DB
	lockedBalance := cModels.LockBalance{}
	if err = db.Where("account_id =? AND asset_id =?", usr.LockAccount.ID, assetId).First(&lockedBalance).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			btlLog.CUST.Error(err.Error())
			return ServiceError, 0, 0, 0
		}
		locked = 0
		err = nil
	}
	locked = lockedBalance.Amount
	tag1 = lockedBalance.Tag1

	assetBalance := cModels.AccountBalance{}
	if err = db.Where("account_id =? AND asset_id =?", usr.Account.ID, assetId).First(&assetBalance).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			btlLog.CUST.Error(err.Error())
			return ServiceError, 0, 0, 0
		}
		unlock = 0
		err = nil
	}
	unlock = assetBalance.Amount
	return
}

func LockAsset(usr *caccount.UserInfo, lockedId string, assetId string, amount float64, tag int) error {

	tx := middleware.DB.Begin()
	defer tx.Rollback()
	var err error

	lockedBalance := cModels.LockBalance{}
	if err = tx.Where("account_id =? AND asset_id =?", usr.LockAccount.ID, assetId).First(&lockedBalance).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			btlLog.CUST.Error(err.Error())
			return ServiceError
		}

		lockedBalance.AssetId = assetId
		lockedBalance.AccountID = usr.LockAccount.ID
		lockedBalance.Amount = 0
	}
	lockedBalance.Amount += amount
	switch tag {
	case 0:
	case 1:
		lockedBalance.Tag1 += amount
	default:
		return fmt.Errorf("invalid tag")
	}
	if err = tx.Save(&lockedBalance).Error; err != nil {
		btlLog.CUST.Error(err.Error())
		return ServiceError
	}

	lockBill := cModels.LockBill{
		AccountID: usr.LockAccount.ID,
		AssetId:   assetId,
		Amount:    amount,
		LockId:    lockedId,
		BillType:  cModels.LockBillTypeLock,
	}
	if err = tx.Create(&lockBill).Error; err != nil {
		var mySQLErr *mysql.MySQLError
		if errors.As(err, &mySQLErr) {
			if mySQLErr.Number == 1062 {
				return RepeatedLockId
			}
		}
		btlLog.CUST.Error(err.Error())
		return ServiceError
	}
	Invoice := InvoiceLocked

	balanceBill := models.Balance{
		AccountId:   usr.Account.ID,
		BillType:    models.BiLLTypeLock,
		Away:        models.AWAY_OUT,
		Amount:      amount,
		Unit:        models.UNIT_ASSET_NORMAL,
		ServerFee:   0,
		AssetId:     &assetId,
		Invoice:     &Invoice,
		PaymentHash: &lockedId,
		State:       models.STATE_SUCCESS,
		TypeExt: &models.BalanceTypeExt{
			Type: models.BTExtLocked,
		},
	}
	if err = tx.Create(&balanceBill).Error; err != nil {
		btlLog.CUST.Error(err.Error())
		return ServiceError
	}
	_, err = custodyBalance.LessAssetBalance(tx, usr, balanceBill.Amount, balanceBill.ID, *balanceBill.AssetId, cModels.ChangeTypeLock)
	if err != nil {
		return err
	}
	tx.Commit()
	return nil
}

func UnlockAsset(usr *caccount.UserInfo, lockedId string, assetId string, amount float64, tag int) error {
	tx := middleware.DB.Begin()
	defer tx.Rollback()
	var err error

	lockedBalance := cModels.LockBalance{}
	if err = tx.Where("account_id =? AND asset_id =?", usr.LockAccount.ID, assetId).First(&lockedBalance).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			btlLog.CUST.Error(err.Error())
			return ServiceError
		}
		lockedBalance.Amount = 0
	}

	if tag == 0 {
		if lockedBalance.Amount < amount {
			return NoEnoughBalance
		}
		if (lockedBalance.Amount - lockedBalance.Tag1) < amount {
			return fmt.Errorf("%w,have  %f is disable unlock", NoEnoughBalance, lockedBalance.Tag1)
		}

		lockedBalance.Amount -= amount
	} else if tag == 1 {
		if lockedBalance.Tag1 < amount {
			return fmt.Errorf("%w,have  %f ", NoEnoughBalance, lockedBalance.Tag1)
		}

		lockedBalance.Amount -= amount
		lockedBalance.Tag1 -= amount
	} else {
		return fmt.Errorf("invalid tag")
	}

	if err = tx.Save(&lockedBalance).Error; err != nil {
		btlLog.CUST.Error(err.Error())
		return ServiceError
	}

	unlockBill := cModels.LockBill{
		AccountID: usr.LockAccount.ID,
		AssetId:   assetId,
		Amount:    amount,
		LockId:    lockedId,
		BillType:  cModels.LockBillTypeUnlock,
	}
	if err = tx.Create(&unlockBill).Error; err != nil {
		var mySQLErr *mysql.MySQLError
		if errors.As(err, &mySQLErr) {
			if mySQLErr.Number == 1062 {
				return RepeatedLockId
			}
		}
		btlLog.CUST.Error(err.Error())
		return ServiceError
	}

	Invoice := InvoiceUnlocked

	balanceBill := models.Balance{
		AccountId:   usr.Account.ID,
		BillType:    models.BiLLTypeLock,
		Away:        models.AWAY_IN,
		Amount:      amount,
		Unit:        models.UNIT_ASSET_NORMAL,
		ServerFee:   0,
		AssetId:     &assetId,
		Invoice:     &Invoice,
		PaymentHash: &lockedId,
		State:       models.STATE_SUCCESS,
		TypeExt: &models.BalanceTypeExt{
			Type: models.BTExtLocked,
		},
	}
	if err = tx.Create(&balanceBill).Error; err != nil {
		btlLog.CUST.Error(err.Error())
		return ServiceError
	}

	_, err = custodyBalance.AddAssetBalance(tx, usr, balanceBill.Amount, balanceBill.ID, *balanceBill.AssetId, cModels.ChangeTypeUnlock)
	if err != nil {
		return err
	}
	tx.Commit()
	return nil
}

func transferLockedAsset(usr *caccount.UserInfo, lockedId string, assetId string, amount float64, toUser *caccount.UserInfo, tag int) error {
	tx := middleware.DB.Begin()
	defer tx.Rollback()

	var err error

	lockedBalance := cModels.LockBalance{}
	if err = tx.Where("account_id =? AND asset_id =?", usr.LockAccount.ID, assetId).First(&lockedBalance).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			btlLog.CUST.Error(err.Error())
			return ServiceError
		}
		lockedBalance.Amount = 0
	}
	if tag == 0 {
		if lockedBalance.Amount < amount {
			return NoEnoughBalance
		}
		if (lockedBalance.Amount - lockedBalance.Tag1) < amount {
			return fmt.Errorf("%w,have  %f is disable unlock", NoEnoughBalance, lockedBalance.Tag1)
		}

		lockedBalance.Amount -= amount
	} else if tag == 1 {
		if lockedBalance.Tag1 < amount {
			return fmt.Errorf("%w,have  %f ", NoEnoughBalance, lockedBalance.Tag1)
		}

		lockedBalance.Amount -= amount
		lockedBalance.Tag1 -= amount
	} else {
		return fmt.Errorf("invalid tag")
	}
	if err = tx.Save(&lockedBalance).Error; err != nil {
		btlLog.CUST.Error(err.Error())
		return ServiceError
	}

	transferBill := cModels.LockBill{
		AccountID: usr.LockAccount.ID,
		AssetId:   assetId,
		Amount:    amount,
		LockId:    lockedId,
		BillType:  cModels.LockBillTypeTransferByLockAsset,
	}
	if err = tx.Create(&transferBill).Error; err != nil {
		var mySQLErr *mysql.MySQLError
		if errors.As(err, &mySQLErr) {
			if mySQLErr.Number == 1062 {
				return RepeatedLockId
			}
		}
		btlLog.CUST.Error(err.Error())
		return ServiceError
	}

	BillExt := cModels.LockBillExt{
		BillId:     transferBill.ID,
		LockId:     lockedId,
		PayAccType: cModels.LockBillExtPayAccTypeLock,
		PayAccId:   usr.LockAccount.ID,
		RevAccId:   toUser.Account.ID,
		Amount:     amount,
		AssetId:    assetId,
		Status:     cModels.LockBillExtStatusSuccess,
	}
	if err = tx.Create(&BillExt).Error; err != nil {
		btlLog.CUST.Error(err.Error())
		return ServiceError
	}

	invoice := InvoicePendingOderReceive
	if usr.User.Username == FeeNpubkey {
		invoice = InvoicePendingOderAward
	}

	balanceBill := models.Balance{
		AccountId:   toUser.Account.ID,
		BillType:    models.BillTypePendingOder,
		Away:        models.AWAY_IN,
		Amount:      amount,
		Unit:        models.UNIT_ASSET_NORMAL,
		ServerFee:   0,
		AssetId:     &assetId,
		Invoice:     &invoice,
		PaymentHash: &lockedId,
		State:       models.STATE_SUCCESS,
		TypeExt: &models.BalanceTypeExt{
			Type: models.BTExtLockedTransfer,
		},
	}
	if err = tx.Create(&balanceBill).Error; err != nil {
		btlLog.CUST.Error(err.Error())
		return ServiceError
	}

	_, err = custodyBalance.AddAssetBalance(tx, toUser, balanceBill.Amount, balanceBill.ID, *balanceBill.AssetId, cModels.ChangeTypeLockedTransfer)
	if err != nil {
		return err
	}

	tx.Commit()
	return nil
}

func transferAsset(usr *caccount.UserInfo, lockedId string, assetId string, amount float64, toUser *caccount.UserInfo) error {
	tx := middleware.DB.Begin()
	defer tx.Rollback()

	var err error

	transferBill := cModels.LockBill{
		AccountID: usr.LockAccount.ID,
		LockId:    lockedId,
		AssetId:   assetId,
		Amount:    amount,
		BillType:  cModels.LockBillTypeTransferByUnlockAsset,
	}
	if err = tx.Create(&transferBill).Error; err != nil {
		var mySQLErr *mysql.MySQLError
		if errors.As(err, &mySQLErr) {
			if mySQLErr.Number == 1062 {
				return RepeatedLockId
			}
		}
		btlLog.CUST.Error(err.Error())
		return ServiceError
	}

	BillExt := cModels.LockBillExt{
		BillId:     transferBill.ID,
		LockId:     lockedId,
		PayAccType: cModels.LockBillExtPayAccTypeUnlock,
		PayAccId:   usr.Account.ID,
		RevAccId:   toUser.Account.ID,
		Amount:     amount,
		AssetId:    assetId,
		Status:     cModels.LockBillExtStatusInit,
	}
	if err = tx.Create(&BillExt).Error; err != nil {
		var mySQLErr *mysql.MySQLError
		if errors.As(err, &mySQLErr) {
			if mySQLErr.Number == 1062 {
				return RepeatedLockId
			}
		}
		btlLog.CUST.Error(err.Error())
		return ServiceError
	}

	payInvoice := InvoicePendingOderPay
	if usr.User.Username == FeeNpubkey {
		payInvoice = InvoicePendingOderAward
	}

	balanceBill := models.Balance{
		AccountId:   usr.Account.ID,
		BillType:    models.BillTypePendingOder,
		Away:        models.AWAY_OUT,
		Amount:      amount,
		Unit:        models.UNIT_ASSET_NORMAL,
		ServerFee:   0,
		AssetId:     &assetId,
		Invoice:     &payInvoice,
		PaymentHash: &lockedId,
		State:       models.STATE_SUCCESS,
		TypeExt: &models.BalanceTypeExt{
			Type: models.BTExtLockedTransfer,
		},
	}
	if err = tx.Create(&balanceBill).Error; err != nil {
		btlLog.CUST.Error(err.Error())
		return ServiceError
	}

	_, err = custodyBalance.LessAssetBalance(tx, usr, balanceBill.Amount, balanceBill.ID, *balanceBill.AssetId, cModels.ChangeTypeLockedTransfer)
	if err != nil {
		return err
	}
	tx.Commit()

	txRev := middleware.DB.Begin()
	defer txRev.Rollback()

	recInvoice := InvoicePendingOderReceive
	if usr.User.Username == FeeNpubkey {
		recInvoice = InvoicePendingOderAward
	}

	balanceBillRev := models.Balance{
		AccountId:   toUser.Account.ID,
		BillType:    models.BillTypePendingOder,
		Away:        models.AWAY_IN,
		Amount:      amount,
		Unit:        models.UNIT_ASSET_NORMAL,
		ServerFee:   0,
		AssetId:     &assetId,
		Invoice:     &recInvoice,
		PaymentHash: &lockedId,
		State:       models.STATE_SUCCESS,
		TypeExt: &models.BalanceTypeExt{
			Type: models.BTExtLockedTransfer,
		},
	}
	if err = txRev.Create(&balanceBillRev).Error; err != nil {
		btlLog.CUST.Error(err.Error())
		return ServiceError
	}

	BillExt.Status = cModels.LockBillExtStatusSuccess
	if err = txRev.Save(&BillExt).Error; err != nil {
		btlLog.CUST.Error(err.Error())
		return ServiceError
	}

	_, err = custodyBalance.AddAssetBalance(txRev, toUser, balanceBillRev.Amount, balanceBillRev.ID, *balanceBillRev.AssetId, cModels.ChangeTypeLockedTransfer)
	if err != nil {
		return err
	}
	txRev.Commit()
	return nil
}
