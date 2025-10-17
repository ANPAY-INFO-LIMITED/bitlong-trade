package custodyModels

import "gorm.io/gorm"

type LockBalance struct {
	gorm.Model
	AccountID uint    `gorm:"column:account_id;type:bigint unsigned;uniqueIndex:idx_account_id_asset_id" json:"accountId"`
	AssetId   string  `gorm:"column:asset_id;type:varchar(128);uniqueIndex:idx_account_id_asset_id" json:"assetId"`
	Amount    float64 `gorm:"type:decimal(25,2);column:amount" json:"amount"`
	Tag1      float64 `gorm:"type:decimal(25,2);column:Tag1" json:"Tag1"`
}

func (LockBalance) TableName() string {
	return "user_lock_balance"
}
