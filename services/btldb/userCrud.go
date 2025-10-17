package btldb

import (
	"trade/middleware"
	"trade/models"
)

func CreateUser(user *models.User) error {
	return middleware.DB.Create(user).Error
}

func ReadUser(id uint) (*models.User, error) {
	var user models.User
	err := middleware.DB.First(&user, id).Error
	return &user, err
}

func ReadAllUser() (*[]models.User, error) {
	var users []models.User
	err := middleware.DB.Find(&users).Error
	return &users, err
}

func ReadUserByUsername(username string) (*models.User, error) {
	var user models.User
	err := middleware.DB.Where("user_name = ?", username).First(&user).Error
	return &user, err
}

func UpdateUser(user *models.User) error {
	return middleware.DB.Save(user).Error
}

func DeleteUser(id uint) error {
	var user models.User
	return middleware.DB.Delete(&user, id).Error
}
