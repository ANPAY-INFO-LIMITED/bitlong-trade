package btldb

import (
	"gorm.io/gorm"
	"sync"
	"trade/middleware"
	"trade/models/custodyModels"
)

var payOutsideMutex sync.Mutex

func CreatePayOutside(pay *custodyModels.PayOutside) error {
	payOutsideMutex.Lock()
	defer payOutsideMutex.Unlock()
	return middleware.DB.Create(pay).Error
}

func ReadPayOutside(id uint) (*custodyModels.PayOutside, error) {
	var pay custodyModels.PayOutside
	err := middleware.DB.First(&pay, id).Error
	return &pay, err
}

func LoadPendingOutsides() (*[]custodyModels.PayOutside, error) {
	var pay []custodyModels.PayOutside
	err := middleware.DB.Where("status =?", custodyModels.PayOutsideStatusPending).Find(&pay).Error
	return &pay, err
}

func UpdatePayOutside(tx *gorm.DB, pay *custodyModels.PayOutside) error {
	payOutsideMutex.Lock()
	defer payOutsideMutex.Unlock()
	return tx.Save(pay).Error
}

func DeletePayOutside(id uint) error {
	var pay custodyModels.PayOutside
	return middleware.DB.Delete(&pay, id).Error
}
