package models

import (
	"gorm.io/gorm"
	"trade/models/custodyModels/store"
)

type Account struct {
	gorm.Model
	UserId   uint          `gorm:"column:user_id;type:bigint unsigned;uniqueIndex:idx_user_id" json:"userId"`
	UserName string        `gorm:"column:user_name;type:varchar(100);index:idx_user_name" json:"userName"`
	Status   AccountStatus `gorm:"column:status;type:smallint" json:"status"`
	Type     AccountType   `gorm:"column:type;type:smallint;default:0" json:"type"`
	Store    *store.Store  `gorm:"foreignKey:AccountId;constraint:OnDelete:CASCADE" json:"omitempty"`
}

func (Account) TableName() string {
	return "user_account"
}

type AccountStatus int16

const (
	AccountStatusDisable AccountStatus = 0
	AccountStatusEnable  AccountStatus = 1
)

type AccountType uint8

const (
	NormalAccount      AccountType = 0
	GameReceiveAccount AccountType = 41
)
