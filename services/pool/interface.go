package pool

import (
	"gorm.io/gorm"
	"math/big"
	"trade/models/custodyModels/pAccount"
	"trade/services/custodyAccount/poolAccount"
)

func CreatePoolAccount(tx *gorm.DB, pairId uint, allowTokens []string) (err error) {
	transTokens := make([]string, 0)
	for _, token := range allowTokens {
		if token == TokenSatTag {
			transTokens = append(transTokens, "00")
		} else {
			transTokens = append(transTokens, token)
		}
	}

	return poolAccount.CreatePoolAccount(tx, pairId, 0, transTokens)
}

// @Description: token is the asset_id or the "sat"
func PoolAccountTransfer(tx *gorm.DB, pairId uint, username string, token string, _amount *big.Int, transferDescription string) (recordId uint, err error) {
	if token == TokenSatTag {
		token = "00"
	}
	return poolAccount.PAccountToUserPay(tx, username, pairId, 0, token, _amount, transferDescription)
}

func TransferToPoolAccount(tx *gorm.DB, username string, pairId uint, token string, _amount *big.Int, transferDescription string) (recordId uint, err error) {
	if token == TokenSatTag {
		token = "00"
	}
	return poolAccount.UserPayToPAccount(tx, pairId, 0, username, token, _amount, transferDescription)
}

func GetPoolAccountRecords(pairId uint, limit int, offset int) (records *[]pAccount.PAccountBill, err error) {
	return poolAccount.GetAccountRecords(pairId, 0, limit, offset)
}
func GetPoolAccountRecordsCount(pairId uint) (count int64, err error) {
	return poolAccount.GetAccountRecordCount(pairId, 0)
}

func GetPoolAccountInfo(pairId uint) (info *poolAccount.PAccountInfo, err error) {
	return poolAccount.GetPoolAccountInfo(pairId, 0)
}

// TODO 6.LockPoolAccount
func LockPoolAccount(tx *gorm.DB, pairId uint) (err error) {
	return poolAccount.LockPoolAccount(tx, pairId, 0)
}

// TODO 7.UnLockPoolAccount
func UnLockPoolAccount(tx *gorm.DB, pairId uint) (err error) {
	return poolAccount.UnlockPoolAccount(tx, pairId, 0)
}

// @Note: Transfer Sats only
func TransferWithdrawReward(username string, _amount *big.Int, transferDescription string) (recordId uint, err error) {
	return poolAccount.AwardSat(username, _amount, transferDescription)
}
