package assetBalanceBackend

import (
	"time"
	"trade/middleware"
	"trade/utils"
)

type AssetBalanceInfo struct {
	ID         uint   `json:"id"`
	CreateTime int64  `json:"create_time"`
	UpdateTime int64  `json:"update_time"`
	Balance    string `json:"balance"`
	Username   string `json:"username"`
}

type AssetBalanceInfoScan struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Balance   string    `json:"balance"`
	Username  string    `json:"username"`
}

func AssetBalanceInfoScanToAssetBalanceInfo(assetBalanceInfoScan AssetBalanceInfoScan) (assetBalanceInfo AssetBalanceInfo) {
	return AssetBalanceInfo{
		ID:         assetBalanceInfoScan.ID,
		CreateTime: assetBalanceInfoScan.CreatedAt.Unix(),
		UpdateTime: assetBalanceInfoScan.UpdatedAt.Unix(),
		Balance:    assetBalanceInfoScan.Balance,
		Username:   assetBalanceInfoScan.Username,
	}
}

func AssetBalanceInfoScansToAssetBalanceInfos(assetBalanceInfoScans []AssetBalanceInfoScan) (assetBalanceInfos []AssetBalanceInfo) {
	assetBalanceInfos = make([]AssetBalanceInfo, len(assetBalanceInfoScans))
	for i, assetBalanceInfoScan := range assetBalanceInfoScans {
		assetBalanceInfos[i] = AssetBalanceInfoScanToAssetBalanceInfo(assetBalanceInfoScan)
	}
	return assetBalanceInfos
}

func GetAssetBalanceLimitAndOffset(assetId string, limit int, offset int) (assetBalanceInfos []AssetBalanceInfo, err error) {
	tx := middleware.DB

	var assetBalanceInfoScans []AssetBalanceInfoScan

	err = tx.Table("asset_balances").
		Select("id,created_at,updated_at,balance,username").
		Where("asset_id = ?", assetId).
		Limit(limit).
		Offset(offset).
		Order("balance desc").
		Scan(&assetBalanceInfoScans).
		Error
	if err != nil {
		return nil, utils.AppendErrorInfo(err, "select asset_balances")
	}
	return AssetBalanceInfoScansToAssetBalanceInfos(assetBalanceInfoScans), nil
}

func GetAssetBalanceCount(assetId string) (count int64, err error) {
	tx := middleware.DB

	err = tx.Table("asset_balances").
		Where("asset_id = ?", assetId).
		Count(&count).
		Error
	if err != nil {
		return 0, utils.AppendErrorInfo(err, "select asset_balances count")
	}
	return count, nil
}

func QueryAssetBalanceInfoByUsername(assetId string, username string) (assetBalanceInfos []AssetBalanceInfo, err error) {
	tx := middleware.DB

	var assetBalanceInfoScans []AssetBalanceInfoScan

	err = tx.Table("asset_balances").
		Select("id,created_at,updated_at,balance,username").
		Where("asset_id = ? and username = ?", assetId, username).
		Scan(&assetBalanceInfos).
		Error
	if err != nil {
		return nil, utils.AppendErrorInfo(err, "select asset_balances")
	}
	return AssetBalanceInfoScansToAssetBalanceInfos(assetBalanceInfoScans), nil
}

func QueryAllAssetBalanceAssetIds() (assetIds []string, err error) {
	tx := middleware.DB

	err = tx.Table("asset_balances").
		Select("asset_id").
		Group("asset_id").
		Pluck("asset_id", &assetIds).
		Error

	if err != nil {
		return nil, utils.AppendErrorInfo(err, "select asset_balances")
	}
	return assetIds, nil
}
