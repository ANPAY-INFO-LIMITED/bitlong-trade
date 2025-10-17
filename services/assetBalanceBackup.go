package services

import (
	"trade/models"
	"trade/services/btldb"
)

func GetAssetBalanceBackup(username string) *models.AssetBalanceBackup {
	assetBalanceBackup, err := btldb.ReadAssetBalanceBackupByUsername(username)
	if err != nil {

		return new(models.AssetBalanceBackup)
	}
	return assetBalanceBackup
}

func UpdateAssetBalanceBackup(username string, hash string) error {
	backup, err := btldb.ReadAssetBalanceBackupByUsername(username)
	if err != nil {

		return btldb.CreateAssetBalanceBackup(&models.AssetBalanceBackup{
			Username: username,
			Hash:     hash,
		})
	}
	backup.Hash = hash
	return btldb.UpdateAssetBalanceBackup(backup)
}
