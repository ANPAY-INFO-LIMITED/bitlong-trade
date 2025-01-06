package pAccount

import "gorm.io/gorm"

type PoolAccount struct {
	gorm.Model
	PairId uint `gorm:"column:pair_id;uniqueIndex:idx_pairId_type;not null"`
	Type   uint `gorm:"column:type;uniqueIndex:idx_pairId_type;not null"`
	Status uint `gorm:"column:status"`
}

func (PoolAccount) TableName() string {
	return "custody_pool_accounts"
}
