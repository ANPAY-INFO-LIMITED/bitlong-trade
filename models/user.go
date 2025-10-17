package models

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username          string `gorm:"unique;column:user_name;type:varchar(255)" json:"userName"`
	Password          string `gorm:"column:password" json:"password"`
	Status            int16  `gorm:"column:status;type:smallint" json:"status"`
	RecentIpAddresses string `json:"recent_ip_addresses" gorm:"type:varchar(255)"`
	RecentLoginTime   int    `json:"recent_login_time"`

	WeakButFastPass string `json:"weak_but_fast_pass" gorm:"type:varchar(255)"`
}

func (User) TableName() string {
	return "user"
}
