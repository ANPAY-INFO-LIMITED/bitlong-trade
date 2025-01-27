package custodyBtc

import (
	"errors"
	"github.com/go-sql-driver/mysql"
	"trade/btlLog"
	"trade/middleware"
	"trade/models"
	"trade/models/custodyModels"
	caccount "trade/services/custodyAccount/account"
)

func PutInAward(user *caccount.UserInfo, _ string, amount int, memo *string, lockedId string) (*models.AccountAward, error) {
	var err error
	tx, back := middleware.GetTx()
	defer back()
	// Build a database Balance
	ba := models.Balance{
		TypeExt: &models.BalanceTypeExt{
			Type: models.BTExtAward,
		},
	}
	ba.AccountId = user.Account.ID
	ba.Amount = float64(amount)
	ba.Unit = models.UNIT_SATOSHIS
	ba.BillType = models.BillTypeAwardSat
	ba.Away = models.AWAY_IN
	ba.State = models.STATE_SUCCESS
	invoiceType := "award"
	ba.Invoice = nil
	ba.PaymentHash = memo
	ba.ServerFee = 0
	ba.Invoice = &invoiceType
	if err = tx.Create(&ba).Error; err != nil {
		btlLog.CUST.Error(err.Error())
		return nil, err
	}
	// Build a database AccountAward
	award := models.AccountAward{
		AccountID: user.Account.ID,
		AssetId:   "00",
		Amount:    float64(amount),
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
				return nil, errors.New("RepeatedLockId")
			}
		}
		return nil, errors.New("ServiceError")
	}
	// Build a database  AccountAwardExt
	awardExt := models.AccountAwardExt{
		BalanceId: ba.ID,
		AwardId:   award.ID,
	}
	if err = tx.Create(&awardExt).Error; err != nil {
		btlLog.CUST.Error(err.Error())
		return nil, err
	}

	if amount < 0 || amount > 10000000 {
		return nil, errors.New("award amount is error")
	}
	// Add btc balance
	_, err = AddBtcBalance(tx, user, ba.Amount, ba.ID, custodyModels.ChangeTypeAward)
	if err != nil {
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
	payba.Amount = ba.Amount
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
		BalanceId: payba.ID,
		AwardId:   award.ID,
	}
	if err = tx.Create(&payAwardExt).Error; err != nil {
		btlLog.CUST.Error(err.Error())
		return nil, err
	}
	_, err = LessBtcBalance(tx, adminUsr, payba.Amount, payba.ID, custodyModels.ChangeTypeOfferAward)
	if err != nil {
		btlLog.CUST.Error(err.Error())
		return nil, err
	}
	tx.Commit()
	return &award, nil
}
