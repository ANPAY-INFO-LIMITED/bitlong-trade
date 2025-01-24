package custodyAssets

import (
	"errors"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
	"sync"
	"trade/btlLog"
	"trade/middleware"
	"trade/models"
	"trade/models/custodyModels"
	caccount "trade/services/custodyAccount/account"
)

var (
	ServerBusy    = errors.New("seriver is busy, please try again later")
	NoAwardType   = fmt.Errorf("no award type")
	AssetIdLock   = fmt.Errorf("award is lock")
	NoEnoughAward = fmt.Errorf("not enough award")
)
var (
	AwardLock = sync.Mutex{}
)

func PutInAward(user *caccount.UserInfo, AssetId string, amount int, memo *string, lockedId string) (*models.AccountAward, error) {
	tx, back := middleware.GetTx()
	if tx == nil {
		return nil, ServerBusy
	}
	defer back()
	// Check if the asset is award type
	var in models.AwardInventory
	err := tx.Where("asset_Id =? ", AssetId).First(&in).Error
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
	if in.Amount < float64(amount) {
		return nil, NoEnoughAward
	}

	AwardLock.Lock()
	defer AwardLock.Unlock()

	// Update the award inventory
	in.Amount -= float64(amount)
	err = tx.Save(&in).Error
	if err != nil {
		btlLog.CUST.Error("err:%v", err)
		return nil, ServerBusy
	}

	// Build a database balance
	ba := models.Balance{
		TypeExt: &models.BalanceTypeExt{
			Type: models.BTExtAward,
		},
	}
	ba.AccountId = user.Account.ID
	ba.Amount = float64(amount)
	ba.Unit = models.UNIT_ASSET_NORMAL
	ba.BillType = models.BillTypeAwardAsset
	ba.Away = models.AWAY_IN
	ba.AssetId = &AssetId
	ba.State = models.STATE_SUCCESS
	invoiceType := "award"
	ba.Invoice = nil
	ba.PaymentHash = memo
	ba.ServerFee = 0
	ba.Invoice = &invoiceType
	// Update the database
	dbErr := tx.Create(&ba).Error
	if dbErr != nil {
		btlLog.CUST.Error(dbErr.Error())
		return nil, ServerBusy
	}
	// Build a database AccountAward
	award := models.AccountAward{
		AccountID: user.Account.ID,
		AssetId:   AssetId,
		Amount:    float64(amount),
		Memo:      memo,
	}
	err = tx.Create(&award).Error
	if err != nil {
		btlLog.CUST.Error(err.Error())
		return nil, ServerBusy
	}

	_, err = AddAssetBalance(tx, user, ba.Amount, ba.ID, *ba.AssetId, custodyModels.ChangeTypeAward)
	if err != nil {
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
	// Build a database AccountAwardExt
	awardExt := models.AccountAwardExt{
		BalanceId: ba.ID,
		AwardId:   award.ID,
	}
	err = tx.Create(&awardExt).Error
	if err != nil {
		btlLog.CUST.Error(err.Error())
		return nil, ServerBusy
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
	payba.AssetId = &AssetId
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
	_, err = LessAssetBalance(tx, adminUsr, payba.Amount, payba.ID, AssetId, custodyModels.ChangeTypeOfferAward)
	if err != nil {
		btlLog.CUST.Error(err.Error())
		return nil, err
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		btlLog.CUST.Error("award failed,not commit:%v", err)
		return nil, ServerBusy
	}
	return &award, nil
}
