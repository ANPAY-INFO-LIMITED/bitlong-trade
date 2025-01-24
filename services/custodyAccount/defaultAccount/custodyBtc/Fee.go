package custodyBtc

import (
	"gorm.io/gorm"
	"trade/models"
	"trade/models/custodyModels"
	"trade/services/custodyAccount/account"
)

func PayFee(Db *gorm.DB, usr *account.UserInfo, amount float64, balanceId uint, invoice, hash *string) error {
	var err error
	_, err = LessBtcBalance(Db, usr, amount, balanceId, custodyModels.ChangeTypeBtcFee)
	if err != nil {
		return err
	}
	var adminUsr *account.UserInfo
	adminUsr, err = account.GetUserInfo("admin")
	if err != nil {
		return err
	}
	ba := models.Balance{}
	ba.AccountId = adminUsr.Account.ID
	ba.Amount = amount
	ba.Unit = models.UNIT_SATOSHIS
	ba.BillType = models.BillTypePaymentFee
	ba.Away = models.AWAY_IN
	ba.Invoice = invoice
	ba.PaymentHash = hash
	ba.State = models.STATE_SUCCESS
	ba.TypeExt = &models.BalanceTypeExt{Type: models.BTEServerFee}
	err = Db.Create(&ba).Error
	if err != nil {
		return err
	}
	_, err = AddBtcBalance(Db, adminUsr, amount, ba.ID, custodyModels.ChangeTypeBtcFee)
	return err
}
