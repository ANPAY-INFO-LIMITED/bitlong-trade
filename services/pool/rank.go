package pool

import (
	"fmt"
	"gorm.io/gorm"
	"trade/middleware"
	"trade/utils"
)

type PoolPairScan struct {
	ID     uint   `json:"id"`
	Token0 string `json:"token0" gorm:"type:varchar(255);uniqueIndex:idx_token_0_token_1"`
	Token1 string `json:"token1" gorm:"type:varchar(255);uniqueIndex:idx_token_0_token_1"`
}

func getAllPoolPairScan() ([]PoolPairScan, error) {
	var poolPairScan []PoolPairScan
	err := middleware.DB.Table("pool_pairs").
		Select("id, token0, token1").
		Scan(&poolPairScan).Error
	if err != nil {
		return nil, utils.AppendErrorInfo(err, "select pool_pairs")
	}
	return poolPairScan, nil
}

func poolPairName(pairId uint, poolType uint, token0, token1 string) string {
	if poolType == PoolTypeFee {
		return "pool" + fmt.Sprintf("%d", pairId) + "FEE" + token0 + token1
	}
	return "pool" + fmt.Sprintf("%d", pairId) + "RSV" + token0 + token1
}

type PoolPairTokenAccountBalance struct {
	gorm.Model
	PairId  uint    `json:"pair_id" gorm:"uniqueIndex:idx_pair_id_token"`
	Token   string  `json:"token" gorm:"type:varchar(255);uniqueIndex:idx_pair_id_token"`
	Balance float64 `json:"balance"`
}

type PoolPairTokenAccountBalanceInfo struct {
	PairId  uint    `json:"pair_id"`
	Token   string  `json:"token"`
	Balance float64 `json:"balance"`
}

func getPoolPairTokenAccountBalanceInfos(pairId uint, token0 string, token1 string) ([]PoolPairTokenAccountBalanceInfo, error) {

	if token0 == TokenSatTag {
		token0 = "00"
	}

	pAccountInfo1, err := GetPoolAccountInfo(pairId, PoolTypeDefault)
	if err != nil {
		return nil, utils.AppendErrorInfo(err, "GetPoolAccountInfo")
	}
	pAccountInfo2, err := GetPoolAccountInfo(pairId, PoolTypeFee)
	if err != nil {
		return nil, utils.AppendErrorInfo(err, "GetPoolAccountInfo")
	}
	var poolPairTokenAccountBalances []PoolPairTokenAccountBalanceInfo

	var tokenMapBalance = make(map[string]float64)

	if pAccountInfo1.Balances != nil {
		for _, balance := range *pAccountInfo1.Balances {
			tokenMapBalance[balance.AssetId] += balance.Balance
		}
	}
	if pAccountInfo2.Balances != nil {
		for _, balance := range *pAccountInfo2.Balances {
			tokenMapBalance[balance.AssetId] += balance.Balance
		}
	}

	poolPairTokenAccountBalances = append(poolPairTokenAccountBalances, PoolPairTokenAccountBalanceInfo{
		PairId:  pairId,
		Token:   token0,
		Balance: tokenMapBalance[token0],
	})

	poolPairTokenAccountBalances = append(poolPairTokenAccountBalances, PoolPairTokenAccountBalanceInfo{
		PairId:  pairId,
		Token:   token1,
		Balance: tokenMapBalance[token1],
	})

	return poolPairTokenAccountBalances, nil
}

func updatePoolPairTokenAccountBalance(tx *gorm.DB, poolPairTokenAccountBalanceInfo PoolPairTokenAccountBalanceInfo) (err error) {
	pairId := poolPairTokenAccountBalanceInfo.PairId
	token := poolPairTokenAccountBalanceInfo.Token
	balance := poolPairTokenAccountBalanceInfo.Balance
	if token == "00" {
		token = TokenSatTag
	}

	var poolPairTokenAccountBalance PoolPairTokenAccountBalance
	err = tx.Table("pool_pair_token_account_balances").
		Where("pair_id = ? AND token = ?", pairId, token).
		First(&poolPairTokenAccountBalance).Error
	if err != nil {

		err = tx.Table("pool_pair_token_account_balances").
			Create(&PoolPairTokenAccountBalance{
				PairId:  pairId,
				Token:   token,
				Balance: balance,
			}).Error
		if err != nil {
			return utils.AppendErrorInfo(err, "create PoolPairTokenAccountBalance")
		}
	}

	err = tx.Table("pool_pair_token_account_balances").
		Where("pair_id = ? AND token = ?", pairId, token).
		Updates(map[string]any{
			"balance": balance,
		}).Error
	if err != nil {
		return utils.AppendErrorInfo(err, "update PoolPairTokenAccountBalance")
	}
	return nil
}

func getAllAndUpdatePoolPairTokenAccountBalance(tx *gorm.DB) (err error) {
	poolPairScan, err := getAllPoolPairScan()
	if err != nil {
		return utils.AppendErrorInfo(err, "getAllPoolPairScan")
	}
	for _, pair := range poolPairScan {
		poolPairTokenAccountBalanceInfos, err := getPoolPairTokenAccountBalanceInfos(pair.ID, pair.Token0, pair.Token1)
		if err != nil {
			return utils.AppendErrorInfo(err, "getPoolPairTokenAccountBalanceInfos")
		}
		for _, poolPairTokenAccountBalanceInfo := range poolPairTokenAccountBalanceInfos {
			err = updatePoolPairTokenAccountBalance(tx, poolPairTokenAccountBalanceInfo)
			if err != nil {
				return utils.AppendErrorInfo(err, "updatePoolPairTokenAccountBalance")
			}
		}
	}
	return nil
}

func UpdateAllPoolPairTokenAccountBalances() error {
	tx := middleware.DB.Begin()
	err := getAllAndUpdatePoolPairTokenAccountBalance(tx)
	if err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

type PoolAccountNameAndBalance struct {
	Name    string  `json:"name"`
	Balance float64 `json:"balance"`
}

type PoolPairTokenAccountBalanceScan struct {
	AccountId string   `gorm:"column:account_id"`
	Balance   float64  `gorm:"column:balance"`
	Type      uint     `gorm:"column:type"`
	PairId    uint     `gorm:"column:pair_id"`
	Token     []string `json:"token"`
}

func GetPoolAccountNameAndBalances(token string) ([]PoolAccountNameAndBalance, error) {
	var poolscans []PoolPairTokenAccountBalanceScan
	err := middleware.DB.Raw(balancesSql, token).Scan(&poolscans).Error
	if err != nil {
		return nil, utils.AppendErrorInfo(err, "select pool_pair_token_account_balances")
	}
	var poolAccountNameAndBalances []PoolAccountNameAndBalance
	for _, scan := range poolscans {
		err = middleware.DB.Raw(getaccountInfoSql, scan.AccountId).Scan(&scan.Token).Error
		if err != nil {
			return nil, utils.AppendErrorInfo(err, "select pool_pair_token_account_balances")
		}
		if len(scan.Token) != 2 {
			continue
		}
		poolAccountNameAndBalances = append(poolAccountNameAndBalances, PoolAccountNameAndBalance{
			Name:    poolPairName(scan.PairId, scan.Type, scan.Token[0], scan.Token[1]),
			Balance: scan.Balance,
		})
	}
	return poolAccountNameAndBalances, nil
}

var balancesSql = `
SELECT pool_account_id as account_id,balance as balance,custody_pool_accounts.type as type,custody_pool_accounts.pair_id as pair_id
FROM custody_pool_account_balances
JOIN custody_pool_accounts ON custody_pool_accounts.id = pool_account_id
WHERE asset_id = ? and balance > 0
`
var getaccountInfoSql = `
SELECT asset_id
FROM custody_pool_account_assetId
where pool_account_id = ?
`

func GetPoolAccountTotalBalance(token string) (float64, error) {
	var err error
	var balances float64
	db := middleware.DB
	err = db.Raw(totalsql, token).Scan(&balances).Error
	if err != nil {
		return 0, utils.AppendErrorInfo(err, "get pool account total balance error")
	}
	return balances, nil
}

var totalsql = `
SELECT COALESCE(sum(balance), 0) as balances
FROM custody_pool_account_balances
where asset_id = ? and balance > 0`

func GetPoolAccountNameAndBalancesCount(token string) (int64, error) {
	var err error
	var count int64
	db := middleware.DB
	err = db.Raw(countsql, token).Scan(&count).Error
	if err != nil {
		return 0, utils.AppendErrorInfo(err, "get pool account count error")
	}
	return count, nil
}

var countsql = `
SELECT COUNT(*) as count
FROM custody_pool_account_balances
where asset_id = ? and balance > 0
`
