package services

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"log"
	"strings"
	"time"
	"trade/config"
	"trade/middleware"
	"trade/models"
	"trade/services/btldb"
)

const fixedSalt = "bitlongwallet7238baee9c2638664"

var aesKey = []byte("YourAESKey32BytesLongForSecurity")

func SplitStringAndVerifyChecksum(extstring string) bool {
	originalString, checksum := spilt(extstring)
	if originalString == "" {
		return false
	}
	if checksum == "" {
		return false
	}
	return verifyChecksumWithSalt(originalString, checksum)
}

func spilt(extstring string) (string, string) {
	parts := strings.Split(extstring, "_e_")
	if len(parts) != 2 {
		return "", ""
	}
	originalString := parts[0]
	checksum := parts[1]
	return originalString, checksum
}

func generateMD5WithSalt(input string) string {
	hasher := md5.New()
	hasher.Write([]byte(input + fixedSalt))
	return hex.EncodeToString(hasher.Sum(nil))
}

func verifyChecksumWithSalt(originalString, checksum string) bool {
	expectedChecksum := generateMD5WithSalt(originalString)
	return checksum == expectedChecksum
}

func ValidAndDecrypt(userName string) (string, error) {
	if !isEncrypted(userName) {
		return "", fmt.Errorf("Username is not encrypted data")
	} else {
		username, err := DecryptAndRestore(userName)
		if err != nil {
			return "", fmt.Errorf("username decryption failed: %v", err)
		}
		return username, nil
	}
}

func Login(req *models.User) (string, error) {

	var username string
	var err error

	if isEncrypted(req.Username) {
		if len(req.Username) <= 0 {
			return "", errors.New("username length negative")
		}

		username, err = DecryptAndRestore(req.Username)
		if err != nil {
			return "", errors.Wrap(err, "DecryptAndRestore")
		}
	} else {
		if config.GetConfig().NetWork == "mainnet" {
			if !isAllNumbers(req.Username) {
				if !(len(req.Username) == 92 || len(req.Username) == 91) {
					return "", errors.New("username length wrong")
				}
			}
		}
		username = req.Username
	}

	var u models.User
	err = middleware.DB.Model(&models.User{}).Where("user_name = ?", username).First(&u).Limit(1).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {

			var password string
			password, err = hashWeakButFastPassword(req.Password)
			if err != nil {
				return "", errors.Wrap(err, "hashWeakButFastPassword")
			}

			err = middleware.DB.Model(&models.User{}).Create(&models.User{
				Username:        username,
				WeakButFastPass: password,
			}).Error

			if err != nil {
				return "", errors.Wrap(err, "middleware.DB.Model(&models.User{}).Create")
			}

		} else {
			return "", errors.Wrap(err, "middleware.DB.Model(&models.User{}).First")
		}
	} else {
		if u.WeakButFastPass == "" {

			err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(req.Password))
			if err != nil {
				return "", errors.Wrap(err, "bcrypt.CompareHashAndPassword")
			}

			var wbfPass string
			wbfPass, err = hashWeakButFastPassword(req.Password)
			if err != nil {
				return "", errors.Wrap(err, "hashWeakButFastPassword")
			}

			err = middleware.DB.Model(&models.User{}).Where("id = ?", u.ID).Update("weak_but_fast_pass", wbfPass).Error
			if err != nil {
				return "", errors.Wrap(err, "middleware.DB.Model(&models.User{}).Update")
			}

		} else {

			err = bcrypt.CompareHashAndPassword([]byte(u.WeakButFastPass), []byte(req.Password))
			if err != nil {
				return "", errors.Wrap(err, "bcrypt.CompareHashAndPassword")
			}

		}
	}

	var token string

	token, err = middleware.GenerateToken(username)
	if err != nil {
		return "", errors.Wrap(err, "middleware.GenerateToken")
	}

	return token, nil
}

func isAllNumbers(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

func isEncrypted(data string) bool {

	if _, err := hex.DecodeString(data); err != nil {
		return false
	}

	if len(data) < 64 {
		return false
	}

	return true
}

func DecryptAndRestore(encryptedData string) (string, error) {
	if !isEncrypted(encryptedData) {
		return encryptedData, nil
	}

	decrypted, err := aesDecrypt(encryptedData)
	if err != nil {
		return "", err
	}

	restored, err := restorePublicKey(decrypted)
	if err != nil {
		return "", err
	}
	return restored, nil
}

func aesDecrypt(encryptedHex string) (string, error) {

	if len(encryptedHex) == 0 {
		return "", fmt.Errorf("empty encrypted data")
	}

	combined, err := hex.DecodeString(encryptedHex)
	if err != nil {
		return "", fmt.Errorf("hex decode error: %v", err)
	}

	if len(combined) < aes.BlockSize {
		return "", fmt.Errorf("invalid ciphertext size")
	}

	iv := combined[:aes.BlockSize]
	ciphertext := combined[aes.BlockSize:]

	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return "", err
	}

	plaintext := make([]byte, len(ciphertext))

	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(plaintext, ciphertext)

	unpadded, err := pkcs7Unpad(plaintext)
	if err != nil {
		return "", err
	}

	return string(unpadded), nil
}

func restorePublicKey(modifiedKey string) (string, error) {

	parts := strings.Split(modifiedKey, "_")

	var result strings.Builder

	for i := 0; i < len(parts); i++ {
		part := parts[i]

		if i > 0 && len(part) > 8 {
			part = part[8:]
		}
		result.WriteString(part)
	}

	return result.String(), nil
}

func pkcs7Unpad(data []byte) ([]byte, error) {
	length := len(data)
	if length == 0 {
		return nil, fmt.Errorf("empty data")
	}

	padding := int(data[length-1])
	if padding > aes.BlockSize || padding == 0 {
		return nil, fmt.Errorf("invalid padding size")
	}

	for i := length - padding; i < length; i++ {
		if data[i] != byte(padding) {
			return nil, fmt.Errorf("invalid padding values")
		}
	}

	return data[:length-padding], nil
}

func ValidateUserAndGenerateToken(creds models.User) (string, error) {
	var (
		username = creds.Username
		err      error
	)
	var user models.User
	result := middleware.DB.Where("user_name = ?", username).First(&user)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return "", errors.New("invalid credentials")
	}
	if !CheckPassword(user.Password, creds.Password) {
		originalString, _ := spilt(creds.Password)
		if originalString != "" {
			password, err := hashPassword(originalString)
			if err != nil {
				return "", err
			}
			user.Password = password

			err = middleware.DB.Model(models.User{}).
				Where("id = ?", user.ID).
				Updates(map[string]any{
					"password":   password,
					"updated_at": time.Now(),
				}).
				Error
			if err != nil {
				return "", err
			}
		}
	}
	token, err := middleware.GenerateToken(username)
	if err != nil {
		return "", err
	}
	return token, nil
}

func ValidateUserAndReChange(creds *models.User) (string, error) {
	var (
		username = creds.Username
		err      error
	)

	if isEncrypted(creds.Username) {
		if len(username) <= 0 {
			return "", fmt.Errorf("username update failed")
		}

		username, err = DecryptAndRestore(creds.Username)
		if err != nil {
			return "", fmt.Errorf("update username decryption failed: %v", err)
		}
		log.Println("update username：" + username)
	} else {
		if config.GetConfig().NetWork != "regtest" {
			if !isAllNumbers(username) {
				if len(username) != len(
					"npub29Z2ncVPR3BRmm9ixwoLF2euPQxKwxXDyPRLtFnH9KepkoudUDq1zBP9MggPF5EMtT3yAfUZ6sEA5tkYm6UJLAHk") {
					return "", fmt.Errorf("username update failed")
				}
			}
		}
	}
	var user models.User
	result := middleware.DB.Where("user_name = ?", username).First(&user).Limit(1)
	if result.Error == nil {
		user.Username = username
		password, err := hashPassword(creds.Password)
		if err != nil {
			return "", err
		}
		user.Password = password
		user.UpdatedAt = time.Now()
		err = btldb.UpdateUser(&user)
		if err != nil {
			return "", err
		}
	}
	if !CheckPassword(user.Password, creds.Password) {
		return "", errors.New("when update invalid credentials")
	}
	token, err := middleware.GenerateToken(username)
	if err != nil {
		return "", err
	}
	creds.Username = username
	return token, nil
}

func (cs *CronService) FiveSecondTask() {
	fmt.Println("5 secs runs")
	log.Println("5 secs runs")
}

func GetUserConfig(username string) (*models.UserConfig, error) {
	db := middleware.DB

	data := struct {
		models.UserConfig
		UserName string `gorm:"column:user_name"`
	}{}
	result := db.Table("user_config").
		Joins(" left join user on user.id = user_config.user_id").
		Where("user_name = ?", username).
		Select("user_config.*, user.user_name").
		Scan(&data)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	if data.Config == "" {
		return nil, nil
	}
	data.User.Username = data.UserName
	return &data.UserConfig, nil
}

func SetUserConfig(username string, config string) int {
	db := middleware.DB
	user := models.User{}
	result := db.Where("user_name = ?", username).First(&user)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return 404
	}
	var userConfig models.UserConfig
	result = db.Where("user_id = ?", user.ID).First(&userConfig)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		userConfig = models.UserConfig{
			UserID: user.ID,
			Config: config,
		}
		db.Create(&userConfig)
	} else if result.Error == nil {
		userConfig.Config = config
		db.Save(&userConfig)
	} else {
		return 500
	}
	return 1
}
