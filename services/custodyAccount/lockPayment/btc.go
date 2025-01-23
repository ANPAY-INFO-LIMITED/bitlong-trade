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
	"trade/services/custodyAccount/defaultAccount/custodyBtc"
)

func GetBtcBalance(usr *caccount.UserInfo) (err error, unlock float64, locked, tag1 float64) {
	db := middleware.DB
	lockedBalance := cModels.LockBalance{}
	if err = db.Where("account_id =? AND asset_id =?", usr.LockAccount.ID, btcId).First(&lockedBalance).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			btlLog.CUST.Error(err.Error())
			return ServiceError, 0, 0, 0
		}
		locked = 0
	}
	locked = lockedBalance.Amount
	tag1 = lockedBalance.Tag1
	e, _ := custodyBtc.NewBtcChannelEvent(usr.User.Username)
	balance, err := e.GetBalance()
	if err != nil {
		return err, 0, 0, 0
	}
	unlock = float64(balance[0].Amount)
	return
}

// LockBTC 冻结BTC
func LockBTC(usr *caccount.UserInfo, lockedId string, amount float64, tag int) error {
	tx, back := middleware.GetTx()
	defer back()

	var err error
	// lock btc
	lockedBalance := cModels.LockBalance{}
	if err = tx.Where("account_id =? AND asset_id =?", usr.LockAccount.ID, btcId).First(&lockedBalance).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			tx.Rollback()
			btlLog.CUST.Error(err.Error())
			return ServiceError
		}
		// Init Balance record
		lockedBalance.AssetId = btcId
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

	// lockBill record
	lockBill := cModels.LockBill{
		AccountID: usr.LockAccount.ID,
		AssetId:   btcId,
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

	BtcId := btcId
	Invoice := InvoiceLocked
	// update user account record
	balanceBill := models.Balance{
		AccountId:   usr.Account.ID,
		BillType:    models.BiLLTypeLock,
		Away:        models.AWAY_OUT,
		Amount:      amount,
		Unit:        models.UNIT_SATOSHIS,
		ServerFee:   0,
		AssetId:     &BtcId,
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

	_, err = custodyBtc.LessBtcBalance(tx, usr, amount, balanceBill.ID, cModels.ChangeTypeLock)
	if err != nil {
		return err
	}

	tx.Commit()
	return nil
}

// UnlockBTC 解冻BTC
func UnlockBTC(usr *caccount.UserInfo, lockedId string, amount float64, tag int) error {
	tx, back := middleware.GetTx()
	defer back()
	var err error

	// check locked balance
	lockedBalance := cModels.LockBalance{}
	if err = tx.Where("account_id =? AND asset_id =?", usr.LockAccount.ID, btcId).First(&lockedBalance).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			btlLog.CUST.Error(err.Error())
			return ServiceError
		}
		lockedBalance.Amount = 0
	}
	// tag 1.0 check awardAmount
	if tag == 0 {
		if lockedBalance.Amount < amount {
			return NoEnoughBalance
		}
		if (lockedBalance.Amount - lockedBalance.Tag1) < amount {
			return fmt.Errorf("%w,have  %f is disable unlock", NoEnoughBalance, lockedBalance.Tag1)
		}
		// update locked balance
		lockedBalance.Amount -= amount
	} else if tag == 1 {
		if lockedBalance.Tag1 < amount {
			return fmt.Errorf("%w,have  %f ", NoEnoughBalance, lockedBalance.Tag1)
		}
		// update locked balance
		lockedBalance.Amount -= amount
		lockedBalance.Tag1 -= amount
	} else {
		return fmt.Errorf("invalid tag")
	}

	if err = tx.Save(&lockedBalance).Error; err != nil {
		btlLog.CUST.Error(err.Error())
		return ServiceError
	}

	// unlockBill record
	unlockBill := cModels.LockBill{
		AccountID: usr.LockAccount.ID,
		AssetId:   btcId,
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

	BtcId := btcId
	Invoice := InvoiceUnlocked

	// update user account record
	balanceBill := models.Balance{
		AccountId:   usr.Account.ID,
		BillType:    models.BiLLTypeLock,
		Away:        models.AWAY_IN,
		Amount:      amount,
		Unit:        models.UNIT_SATOSHIS,
		ServerFee:   0,
		AssetId:     &BtcId,
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

	_, err = custodyBtc.AddBtcBalance(tx, usr, amount, balanceBill.ID, cModels.ChangeTypeUnlock)
	if err != nil {
		return err
	}
	tx.Commit()
	return nil
}

// transferLockedBTC 转账冻结的BTC
func transferLockedBTC(usr *caccount.UserInfo, lockedId string, amount float64, toUser *caccount.UserInfo, tag int) error {
	tx, back := middleware.GetTx()
	defer back()
	BtcId := btcId

	var err error

	// check locked balance
	lockedBalance := cModels.LockBalance{}
	if err = tx.Where("account_id =? AND asset_id =?", usr.LockAccount.ID, btcId).First(&lockedBalance).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			tx.Rollback()
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
		// update locked balance
		lockedBalance.Amount -= amount
	} else if tag == 1 {
		if lockedBalance.Tag1 < amount {
			return fmt.Errorf("%w,have  %f ", NoEnoughBalance, lockedBalance.Tag1)
		}
		// update locked balance
		lockedBalance.Amount -= amount
		lockedBalance.Tag1 -= amount
	} else {
		return fmt.Errorf("invalid tag")
	}
	if err = tx.Save(&lockedBalance).Error; err != nil {
		btlLog.CUST.Error(err.Error())
		return ServiceError
	}

	// unlockBill record
	transferBill := cModels.LockBill{
		AccountID: usr.LockAccount.ID,
		AssetId:   btcId,
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

	// Create transferBTC BillExt
	BillExt := cModels.LockBillExt{
		BillId:     transferBill.ID,
		LockId:     lockedId,
		PayAccType: cModels.LockBillExtPayAccTypeLock,
		PayAccId:   usr.LockAccount.ID,
		RevAccId:   toUser.Account.ID,
		Amount:     amount,
		AssetId:    btcId,
		Status:     cModels.LockBillExtStatusSuccess,
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

	invoice := InvoicePendingOderReceive
	if usr.User.Username == FeeNpubkey {
		invoice = InvoicePendingOderAward
	}

	// update user account record
	balanceBill := models.Balance{
		AccountId:   toUser.Account.ID,
		BillType:    models.BillTypePendingOder,
		Away:        models.AWAY_IN,
		Amount:      amount,
		Unit:        models.UNIT_SATOSHIS,
		ServerFee:   0,
		AssetId:     &BtcId,
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

	_, err = custodyBtc.AddBtcBalance(tx, toUser, amount, balanceBill.ID, cModels.ChangeTypeLockedTransfer)
	if err != nil {
		btlLog.CUST.Error(err.Error())
		return ServiceError
	}
	tx.Commit()

	return nil
}

// transferBTC 转账非冻结的BTC
func transferBTC(usr *caccount.UserInfo, lockedId string, amount float64, toUser *caccount.UserInfo) error {
	BtcId := btcId
	tx, back := middleware.GetTx()
	defer back()

	var err error
	// Create transferBTC Bill
	transferBill := cModels.LockBill{
		AccountID: usr.LockAccount.ID,
		LockId:    lockedId,
		AssetId:   btcId,
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
	// Create transferBTC BillExt
	BillExt := cModels.LockBillExt{
		BillId:     transferBill.ID,
		LockId:     lockedId,
		PayAccType: cModels.LockBillExtPayAccTypeUnlock,
		PayAccId:   usr.Account.ID,
		RevAccId:   toUser.Account.ID,
		Amount:     amount,
		AssetId:    btcId,
		Status:     cModels.LockBillExtStatusInit,
	}
	if err = tx.Create(&BillExt).Error; err != nil {
		btlLog.CUST.Error(err.Error())
		return ServiceError
	}

	payInvoice := InvoicePendingOderPay
	if usr.User.Username == FeeNpubkey {
		payInvoice = InvoicePendingOderAward
	}

	// transfer balance record
	balanceBill := models.Balance{
		AccountId:   usr.Account.ID,
		BillType:    models.BillTypePendingOder,
		Away:        models.AWAY_OUT,
		Amount:      amount,
		Unit:        models.UNIT_SATOSHIS,
		ServerFee:   0,
		AssetId:     &BtcId,
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
	_, err = custodyBtc.LessBtcBalance(tx, usr, amount, balanceBill.ID, cModels.ChangeTypeLockedTransfer)
	if err != nil {
		return err
	}

	recInvoice := InvoicePendingOderReceive
	if usr.User.Username == FeeNpubkey {
		recInvoice = InvoicePendingOderAward
	}

	// update user account record
	balanceBillRev := models.Balance{
		AccountId:   toUser.Account.ID,
		BillType:    models.BillTypePendingOder,
		Away:        models.AWAY_IN,
		Amount:      amount,
		Unit:        models.UNIT_SATOSHIS,
		ServerFee:   0,
		AssetId:     &BtcId,
		Invoice:     &recInvoice,
		PaymentHash: &lockedId,
		State:       models.STATE_SUCCESS,
		TypeExt: &models.BalanceTypeExt{
			Type: models.BTExtLockedTransfer,
		},
	}
	if err = tx.Create(&balanceBillRev).Error; err != nil {
		btlLog.CUST.Error(err.Error())
		return ServiceError
	}
	// update billExt record
	BillExt.Status = cModels.LockBillExtStatusSuccess
	if err = tx.Save(&BillExt).Error; err != nil {
		btlLog.CUST.Error(err.Error())
		return ServiceError
	}

	_, err = custodyBtc.AddBtcBalance(tx, toUser, amount, balanceBillRev.ID, cModels.ChangeTypeLockedTransfer)
	if err != nil {
		btlLog.CUST.Error(err.Error())
		return ServiceError
	}

	tx.Commit()
	return nil
}
