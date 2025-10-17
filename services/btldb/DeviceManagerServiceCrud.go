package btldb

import (
	"trade/middleware"
	"trade/models"
)

func CreateDeviceManager(device *models.DeviceManager) error {
	return middleware.DB.Create(device).Error
}

func ReadDeviceManagerByID(id uint) (*models.DeviceManager, error) {
	var device models.DeviceManager
	err := middleware.DB.First(&device, id).Error
	return &device, err
}

func ReadAllDeviceManagers() (*[]models.DeviceManager, error) {
	var devices []models.DeviceManager
	err := middleware.DB.Find(&devices).Error
	return &devices, err
}

func ReadDeviceManagerByNpubKey(npubKey string) (*models.DeviceManager, error) {
	var device models.DeviceManager
	err := middleware.DB.Where("npub_key = ?", npubKey).First(&device).Error
	return &device, err
}

func UpdateDeviceManager(device *models.DeviceManager) error {
	return middleware.DB.Save(device).Error
}

func DeleteDeviceManager(id uint) error {
	var device models.DeviceManager
	return middleware.DB.Delete(&device, id).Error
}
