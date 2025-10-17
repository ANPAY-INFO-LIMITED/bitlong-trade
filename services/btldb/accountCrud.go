package btldb

import (
	"trade/middleware"
	"trade/models"
)

func CreateAccount(account *models.Account) error {
	return middleware.DB.Create(account).Error
}

func ReadAccount(id uint) (*models.Account, error) {
	var account models.Account
	err := middleware.DB.First(&account, id).Error
	return &account, err
}

func ReadAccountByName(name string) (*models.Account, error) {
	var account models.Account
	err := middleware.DB.Where("user_name =?", name).First(&account).Error
	return &account, err
}

func ReadAccountByUserId(userId uint) (*models.Account, error) {
	var account models.Account
	err := middleware.DB.Where("user_id =?", userId).First(&account).Error
	return &account, err
}

func UpdateAccount(account *models.Account) error {
	return middleware.DB.Save(account).Error
}

func DeleteAccount(id uint) error {
	var account models.Account
	return middleware.DB.Delete(&account, id).Error
}
