package custodyModels

import "gorm.io/gorm"

type VirtualUser struct {
	gorm.Model
	Username string `gorm:"unique;column:user_name;type:varchar(255)" json:"userName"` // 正确地将unique和column选项放在同一个gorm标签内
	Status   int16  `gorm:"column:status;type:smallint" json:"status"`
}

func (VirtualUser) TableName() string {
	return "user"
}
