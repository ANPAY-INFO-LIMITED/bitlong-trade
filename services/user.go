package services

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"log"
	"time"
	"trade/middleware"
	"trade/models"
	"trade/services/btldb"
	"trade/utils"
)

const (
	AdminUploadUserName = "adminUploadUser"
)

func ValidateUser(creds models.User) (string, error) {
	var user models.User
	result := middleware.DB.Where("username = ?", creds.Username).First(&user)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return "", errors.New("invalid credentials")
	}
	if user.Password != creds.Password {
		return "", errors.New("invalid credentials")
	}
	token, err := middleware.GenerateToken(creds.Username)
	if err != nil {
		return "", err
	}
	return token, nil
}

func (cs *CronService) SixSecondTask() {
	fmt.Println("6 secs runs")
	log.Println("6 secs runs")
}

func NameToId(name string) (int, error) {
	user := models.User{}
	err := middleware.DB.Where("user_name = ?", name).First(&user).Error
	return int(user.ID), err
}

func IdToName(id int) (string, error) {
	user, err := btldb.ReadUser(uint(id))
	if err != nil {
		return "", err
	}
	return user.Username, nil
}

func hashPassword(password string) (string, error) {

	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func hashWeakButFastPassword(password string) (string, error) {

	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func CheckPassword(hashedPassword, password string) bool {

	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

func UpdateUserIpByUserId(userId uint, ip string) error {
	user, err := btldb.ReadUser(userId)
	if err != nil {
		return err
	}
	user.RecentIpAddresses = ip
	return btldb.UpdateUser(user)
}

func UpdateUserIpByUsername(username string, ip string) (*models.User, error) {
	user, err := btldb.ReadUserByUsername(username)
	if err != nil {
		return nil, err
	}
	user.RecentIpAddresses = ip
	user.RecentLoginTime = utils.GetTimestamp()
	return user, btldb.UpdateUser(user)
}

func UpdateUserIpByClientIp(c *gin.Context) (string, error) {
	username := c.MustGet("username").(string)
	ip := c.ClientIP()
	user, err := UpdateUserIpByUsername(username, ip)
	return user.RecentIpAddresses, err
}

type UserSimplified struct {
	UpdatedAt         time.Time `json:"updated_at"`
	Username          string    `json:"userName"`
	RecentIpAddresses string    `json:"recent_ip_addresses"`
	RecentLoginTime   int       `json:"recent_login_time"`
}

func UserToUserSimplified(user models.User) UserSimplified {
	return UserSimplified{
		UpdatedAt:         user.UpdatedAt,
		Username:          user.Username,
		RecentIpAddresses: user.RecentIpAddresses,
		RecentLoginTime:   user.RecentLoginTime,
	}
}

func UserSliceToUserSimplifiedSlice(users *[]models.User) *[]UserSimplified {
	if users == nil {
		return nil
	}
	var userSimplified []UserSimplified
	for _, user := range *users {
		userSimplified = append(userSimplified, UserToUserSimplified(user))
	}
	return &userSimplified
}

func GetAllUser() (*[]models.User, error) {
	return btldb.ReadAllUser()
}

func GetAllUserSimplified() (*[]UserSimplified, error) {
	allUsers, err := btldb.ReadAllUser()
	if err != nil {
		return nil, err
	}
	return UserSliceToUserSimplifiedSlice(allUsers), nil
}
