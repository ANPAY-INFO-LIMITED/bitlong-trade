package pool

import (
	"gorm.io/gorm"
	"math/big"
	"trade/models/custodyModels/pAccount"
	"trade/services/custodyAccount/poolAccount"
)

const (
	PoolTypeDefault uint = 0

	PoolTypeFee uint = 1
)

func CreatePoolAccount(tx *gorm.DB, pairId uint, poolType uint, allowTokens []string) (err error) {
	transTokens := make([]string, 0)
	for _, token := range allowTokens {
		if token == TokenSatTag {
			transTokens = append(transTokens, "00")
		} else {
			transTokens = append(transTokens, token)
		}
	}

	return poolAccount.CreatePoolAccount(tx, pairId, poolType, transTokens)
}

func PoolAccountTransfer(tx *gorm.DB, pairId uint, poolType uint, username string, token string, _amount *big.Int, transferDescription string) (recordId uint, err error) {
	if token == TokenSatTag {
		token = "00"
	}
	return poolAccount.PAccountToUserPay(tx, username, pairId, poolType, token, _amount, transferDescription)
}

func TransferToPoolAccount(tx *gorm.DB, username string, pairId uint, poolType uint, token string, _amount *big.Int, transferDescription string) (recordId uint, err error) {
	if token == TokenSatTag {
		token = "00"
	}
	return poolAccount.UserPayToPAccount(tx, pairId, poolType, username, token, _amount, transferDescription)
}

func PoolToPoolPTransfer(tx *gorm.DB, fromPairId uint, fromType uint, toPairId uint, toType uint, token string, _amount *big.Int, transferDescription string) (recordId uint, err error) {
	if token == TokenSatTag {
		token = "00"
	}
	return poolAccount.PAccountToPAccountPay(tx, fromPairId, fromType, toPairId, toType, token, _amount, transferDescription)
}

func GetPoolAccountRecords(pairId uint, poolType uint, limit int, offset int) (records *[]pAccount.PAccountBill, err error) {
	return poolAccount.GetAccountRecords(pairId, poolType, limit, offset)
}

func GetPoolAccountRecordsCount(pairId uint, poolType uint) (count int64, err error) {
	return poolAccount.GetAccountRecordCount(pairId, poolType)
}

func GetPoolAccountInfo(pairId uint, poolType uint) (info *poolAccount.PAccountInfo, err error) {
	return poolAccount.GetPoolAccountInfo(pairId, poolType)
}

func CleanPoolAccount(pairId uint, poolType uint) {
	poolAccount.CleanPoolAccount(pairId, poolType)
}

func LockPoolAccount(tx *gorm.DB, pairId uint, poolType uint) (err error) {
	return poolAccount.LockPoolAccount(tx, pairId, poolType)
}

func UnLockPoolAccount(tx *gorm.DB, pairId uint, poolType uint) (err error) {
	return poolAccount.UnlockPoolAccount(tx, pairId, poolType)
}

func TransferWithdrawReward(username string, _amount *big.Int, transferDescription string) (recordId uint, err error) {
	return poolAccount.AwardSat(username, _amount, transferDescription)
}
