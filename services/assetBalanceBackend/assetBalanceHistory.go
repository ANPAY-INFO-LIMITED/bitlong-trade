package assetBalanceBackend

import (
	"time"
	"trade/middleware"
	"trade/utils"
)

type AssetBalanceHistoryInfoScan struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Balance   string    `json:"balance"`
	Username  string    `json:"username"`
}

type AssetBalanceHistoryInfo struct {
	ID uint `json:"id"`

	Time     int64  `json:"time"`
	Balance  string `json:"balance"`
	Username string `json:"username"`
}

func AssetBalanceHistoryInfoScanToAssetBalanceHistoryInfo(assetBalanceHistoryInfoScan AssetBalanceHistoryInfoScan) (assetBalanceHistoryInfo AssetBalanceHistoryInfo) {
	return AssetBalanceHistoryInfo{
		ID:       assetBalanceHistoryInfoScan.ID,
		Time:     assetBalanceHistoryInfoScan.CreatedAt.Unix(),
		Balance:  assetBalanceHistoryInfoScan.Balance,
		Username: assetBalanceHistoryInfoScan.Username,
	}
}

func AssetBalanceHistoryInfoScansToAssetBalanceHistoryInfos(assetBalanceHistoryInfoScans []AssetBalanceHistoryInfoScan) (assetBalanceHistoryInfos []AssetBalanceHistoryInfo) {
	for _, assetBalanceHistoryInfoScan := range assetBalanceHistoryInfoScans {
		assetBalanceHistoryInfos = append(assetBalanceHistoryInfos, AssetBalanceHistoryInfoScanToAssetBalanceHistoryInfo(assetBalanceHistoryInfoScan))
	}
	return assetBalanceHistoryInfos
}

func GetAssetBalanceHistoryLimitAndOffset(assetId string, limit int, offset int) (AssetBalanceHistoryInfos []AssetBalanceHistoryInfo, err error) {
	tx := middleware.DB

	var assetBalanceHistoryInfoScans []AssetBalanceHistoryInfoScan

	err = tx.Table("asset_balance_histories").
		Select("id,created_at,balance,username").
		Where("asset_id = ?", assetId).
		Limit(limit).
		Offset(offset).
		Order("balance desc").
		Scan(&assetBalanceHistoryInfoScans).
		Error
	if err != nil {
		return nil, utils.AppendErrorInfo(err, "select asset_balance_histories")
	}
	return AssetBalanceHistoryInfoScansToAssetBalanceHistoryInfos(assetBalanceHistoryInfoScans), nil
}

func GetAssetBalanceHistoryCount(assetId string) (count int64, err error) {
	tx := middleware.DB

	err = tx.Table("asset_balance_histories").
		Where("asset_id = ?", assetId).
		Count(&count).
		Error
	if err != nil {
		return 0, utils.AppendErrorInfo(err, "select asset_balance_histories count")
	}
	return count, nil
}

func QueryAssetBalanceHistoryInfoByUsername(assetId string, username string) (AssetBalanceHistoryInfos []AssetBalanceHistoryInfo, err error) {
	tx := middleware.DB

	var assetBalanceHistoryInfoScans []AssetBalanceHistoryInfoScan

	err = tx.Table("asset_balance_histories").
		Select("id,created_at,balance,username").
		Where("asset_id = ? and username = ?", assetId, username).
		Order("id desc").
		Scan(&assetBalanceHistoryInfoScans).
		Error

	if err != nil {
		return nil, utils.AppendErrorInfo(err, "select asset_balance_histories")
	}
	return AssetBalanceHistoryInfoScansToAssetBalanceHistoryInfos(assetBalanceHistoryInfoScans), nil
}

func QueryAllAssetBalanceHistoryAssetIds() (assetIds []string, err error) {
	tx := middleware.DB

	err = tx.Table("asset_balance_histories").
		Select("asset_id").
		Group("asset_id").
		Pluck("asset_id", &assetIds).
		Error

	if err != nil {
		return nil, utils.AppendErrorInfo(err, "select asset_balance_histories")
	}
	return assetIds, nil
}
