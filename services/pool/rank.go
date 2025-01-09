package pool

import (
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

func poolPairName(token0, token1 string) string {
	return "pool" + token0 + token1
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
	// only token0 is possible to be TokenSatTag
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
	//get PoolPairTokenAccountBalance
	var poolPairTokenAccountBalance PoolPairTokenAccountBalance
	err = tx.Table("pool_pair_token_account_balances").
		Where("pair_id = ? AND token = ?", pairId, token).
		First(&poolPairTokenAccountBalance).Error
	if err != nil {
		// no PoolPairTokenAccountBalance, create one
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
	//update PoolPairTokenAccountBalance
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
	Token0  string  `json:"token0"`
	Token1  string  `json:"token1"`
	Balance float64 `json:"balance"`
}

func GetPoolAccountNameAndBalancesCount(token string) (int64, error) {
	if token == "00" {
		token = TokenSatTag
	}
	
	var count int64

	err := middleware.DB.Table("pool_pair_token_account_balances").
		Joins("JOIN pool_pairs ON pool_pairs.id = pool_pair_token_account_balances.pair_id").
		Where("token = ?", token).
		Count(&count).Error

	if err != nil {
		return 0, utils.AppendErrorInfo(err, "select pool_pair_token_account_balances count")
	}

	return count, nil
}

func GetPoolAccountNameAndBalances(token string, limit int, offset int) ([]PoolAccountNameAndBalance, error) {
	if token == "00" {
		token = TokenSatTag
	}
	var poolPairTokenAccountBalanceScans []PoolPairTokenAccountBalanceScan

	err := middleware.DB.Table("pool_pair_token_account_balances").
		Select("pool_pairs.token0, pool_pairs.token1, pool_pair_token_account_balances.balance").
		Joins("JOIN pool_pairs ON pool_pairs.id = pool_pair_token_account_balances.pair_id").
		Where("token = ?", token).
		Order("pool_pair_token_account_balances.balance DESC").
		Limit(limit).
		Offset(offset).
		Scan(&poolPairTokenAccountBalanceScans).Error

	if err != nil {
		return nil, utils.AppendErrorInfo(err, "select pool_pair_token_account_balances")
	}

	var poolAccountNameAndBalances []PoolAccountNameAndBalance
	for _, poolPairTokenAccountBalanceScan := range poolPairTokenAccountBalanceScans {
		poolAccountNameAndBalances = append(poolAccountNameAndBalances, PoolAccountNameAndBalance{
			Name:    poolPairName(poolPairTokenAccountBalanceScan.Token0, poolPairTokenAccountBalanceScan.Token1),
			Balance: poolPairTokenAccountBalanceScan.Balance,
		})
	}

	return poolAccountNameAndBalances, nil
}
