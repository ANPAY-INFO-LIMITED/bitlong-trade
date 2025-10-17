package btldb

import (
	"gorm.io/gorm"
	"sync"
	"trade/middleware"
	"trade/models"
)

var balanceMutex sync.Mutex

func CreateBalance(tx *gorm.DB, balance *models.Balance) error {
	balanceMutex.Lock()
	defer balanceMutex.Unlock()
	return tx.Create(balance).Error
}

func ReadBalance(id uint) (*models.Balance, error) {
	var balance models.Balance
	err := middleware.DB.First(&balance, id).Error
	return &balance, err
}

func UpdateBalance(tx *gorm.DB, balance *models.Balance) error {
	balanceMutex.Lock()
	defer balanceMutex.Unlock()
	return tx.Save(balance).Error
}

func DeleteBalance(id uint) error {
	var balance models.Balance
	return middleware.DB.Delete(&balance, id).Error
}
