package btldb

import (
	"sync"
	"trade/middleware"
	"trade/models/custodyModels"
)

var payOutsideTxMutex sync.Mutex

func CreatePayOutsideTx(pay *custodyModels.PayOutsideTx) error {
	payOutsideTxMutex.Lock()
	defer payOutsideTxMutex.Unlock()
	return middleware.DB.Create(pay).Error
}

func ReadPayOutsideTx(id uint) (*custodyModels.PayOutsideTx, error) {
	var pay custodyModels.PayOutsideTx
	err := middleware.DB.First(&pay, id).Error
	return &pay, err
}

func UpdatePayOutsideTx(pay *custodyModels.PayOutsideTx) error {
	payOutsideTxMutex.Lock()
	defer payOutsideTxMutex.Unlock()
	return middleware.DB.Save(pay).Error
}

func DeletePayOutsideTx(id uint) error {
	var pay custodyModels.PayOutsideTx
	return middleware.DB.Delete(&pay, id).Error
}
