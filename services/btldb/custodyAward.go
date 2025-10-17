package btldb

import (
	"sync"
	"trade/middleware"
	"trade/models"
)

var awardMutex sync.Mutex

func CreateAward(award *models.AccountAward) error {
	awardMutex.Lock()
	defer awardMutex.Unlock()
	return middleware.DB.Create(award).Error
}

func ReadAward(id uint) (*models.AccountAward, error) {
	var award models.AccountAward
	err := middleware.DB.First(&award, id).Error
	return &award, err
}

func UpdateAward(award *models.AccountAward) error {
	payOutsideMutex.Lock()
	defer payOutsideMutex.Unlock()
	return middleware.DB.Save(award).Error
}

func DeleteAward(id uint) error {
	var award models.AccountAward
	return middleware.DB.Delete(&award, id).Error
}
