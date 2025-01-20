package assetMoreInfo

import (
	"trade/middleware"
)

type AssetManagedUtxoInfo struct {
	ID                  uint   `json:"id"`
	OutPoint            string `json:"out_point" gorm:"type:varchar(255)"`
	Time                int    `json:"time"`
	AmtSat              int    `json:"amt_sat"`
	AssetGenesisName    string `json:"asset_genesis_name" gorm:"type:varchar(255)"`
	AssetGenesisAssetID string `json:"asset_genesis_asset_id" gorm:"type:varchar(255)"`
	Amount              int    `json:"amount"`
	Username            string `json:"username" gorm:"type:varchar(255)"`
}

// UTXO

func GetAssetManagedUtxoInfoCount(assetId string) (count int64, err error) {
	err = middleware.DB.Table("asset_managed_utxos").
		Where("asset_genesis_asset_id = ?", assetId).
		Count(&count).
		Error
	return count, err
}

func GetAssetManagedUtxoInfo(assetId string, limit int, offset int) (assetManagedUtxoInfos []AssetManagedUtxoInfo, err error) {
	err = middleware.DB.Table("asset_managed_utxos").
		Select("id, out_point, time, amt_sat, asset_genesis_name, asset_genesis_asset_id, amount, username").
		Where("asset_genesis_asset_id = ?", assetId).
		Order("time desc").
		Limit(limit).
		Offset(offset).
		Scan(&assetManagedUtxoInfos).Error
	return assetManagedUtxoInfos, err
}
