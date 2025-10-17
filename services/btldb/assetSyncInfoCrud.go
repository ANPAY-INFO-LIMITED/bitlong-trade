package btldb

import (
	"trade/middleware"
	"trade/models"
)

func CreateAssetSyncInfo(assetSyncInfo *models.AssetSyncInfo) error {
	return middleware.DB.Create(assetSyncInfo).Error
}

func ReadAssetSyncInfo(id uint) (*models.AssetSyncInfo, error) {
	var assetSyncInfo models.AssetSyncInfo
	err := middleware.DB.First(&assetSyncInfo, id).Error
	return &assetSyncInfo, err
}

func ReadAssetSyncInfoByAssetID(assetID string) (*models.AssetSyncInfo, error) {
	var assetSyncInfo models.AssetSyncInfo
	err := middleware.DB.Where("asset_id =?", assetID).First(&assetSyncInfo).Error
	return &assetSyncInfo, err
}

func UpdateAssetSyncInfo(assetSyncInfo *models.AssetSyncInfo) error {
	return middleware.DB.Save(assetSyncInfo).Error
}

func DeleteAssetSyncInfo(id uint) error {
	var assetSyncInfo models.AssetSyncInfo
	return middleware.DB.Delete(&assetSyncInfo, id).Error
}
