package custodyModels

type Control struct {
	ControlName string `gorm:"primarykey" json:"control_name"`
	Status      bool   `gorm:"not null" json:"status"`
}

func (Control) TableName() string {
	return "user_account_controls"
}
