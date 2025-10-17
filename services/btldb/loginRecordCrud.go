package btldb

import (
	"trade/middleware"
	"trade/models"
)

func CreateLoginRecord(loginRecord *models.LoginRecord) error {
	return middleware.DB.Create(loginRecord).Error
}
