package assetMoreInfo

import (
	"trade/middleware"
)

type AssetBalanceInfo struct {
	Balance  int    `json:"balance"`
	Username string `json:"username"`
}

func GetAssetBalanceInfoCount(assetId string) (count int64, err error) {
	err = middleware.DB.Table("asset_balances").
		Where("asset_id = ? and balance > 0", assetId).
		Count(&count).
		Error
	return count, err
}

func GetAssetBalanceInfo(assetId string, limit int, offset int) (assetBalanceInfos []AssetBalanceInfo, err error) {
	err = middleware.DB.Table("asset_balances").
		Select("balance, username").
		Where("asset_id = ? and balance > 0", assetId).
		Limit(limit).
		Offset(offset).
		Order("balance desc").
		Scan(&assetBalanceInfos).
		Error
	return assetBalanceInfos, err
}
