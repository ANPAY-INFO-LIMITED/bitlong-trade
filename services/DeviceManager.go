package services

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"golang.org/x/crypto/pbkdf2"
	"io"
	"strings"
	"sync"
	"time"
	"trade/middleware"
	"trade/models"
	"trade/services/btldb"
)

func GenerateNonce() (string, error) {

	randomBytes := make([]byte, 16)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}

	hash := sha256.New()
	hash.Write(randomBytes)
	hash.Write([]byte(time.Now().String()))
	return hex.EncodeToString(hash.Sum(nil)), nil
}

var nonceStore = make(map[string]time.Time)

func StoreNonceInRedis(username string, tokenString string) (string, error) {

	userName, err := middleware.RedisGet(tokenString)
	if err != nil && !errors.Is(err, redis.Nil) {
		return "", err
	}
	if userName != "" {

		return tokenString, nil
	}

	if err != nil && !errors.Is(err, redis.Nil) {
		return "", err
	}

	redisSetTimeMinute := 10
	expiration := time.Duration(redisSetTimeMinute) * time.Minute

	err = middleware.RedisSet(tokenString, username+"_nonce", expiration)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func VerifyNonce(nonce string, usernameRef string) bool {

	username, err := middleware.RedisGet(nonce)
	if err != nil || username == "" {
		return false
	}
	if usernameRef+"_nonce" != username {
		return false
	}

	if err := middleware.RedisDel(nonce); err != nil {
		return true
	}
	return true
}

type DeviceIDGenerator struct {
	lastTimestamp int64
	sequence      int64
	mutex         sync.Mutex
}

func NewDeviceIDGenerator() *DeviceIDGenerator {
	return &DeviceIDGenerator{
		lastTimestamp: 0,
		sequence:      0,
		mutex:         sync.Mutex{},
	}
}

func GenerateRandomString(length int) (string, error) {
	bytes := make([]byte, length/2)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
func (g *DeviceIDGenerator) GenerateDeviceID(prefix string, randomLength int) (string, error) {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	currentTimestamp := time.Now().UnixNano() / 1e6

	if currentTimestamp == g.lastTimestamp {
		g.sequence++
	} else {
		g.sequence = 0
		g.lastTimestamp = currentTimestamp
	}

	randomStr, err := GenerateRandomString(randomLength)
	if err != nil {
		return "", err
	}

	var builder strings.Builder

	if prefix != "" {
		builder.WriteString(prefix)
		builder.WriteString("-")
	}

	builder.WriteString(fmt.Sprintf("%d", currentTimestamp))

	builder.WriteString(fmt.Sprintf("%03d", g.sequence))

	builder.WriteString("-")
	builder.WriteString(randomStr)
	return builder.String(), nil
}

func GetDeviceID() (string, error) {
	generator := NewDeviceIDGenerator()
	deviceID, err := generator.GenerateDeviceID("DEV", 8)
	if err != nil {
		return "", err
	}
	return deviceID, nil
}
func generateKeyAndSalt(password []byte) ([]byte, []byte, error) {

	salt := make([]byte, 16)
	_, err := io.ReadFull(rand.Reader, salt)
	if err != nil {
		return nil, nil, err
	}

	key := pbkdf2.Key(password, salt, 10000, 32, sha256.New)

	return key, salt, nil
}
func encrypt(plainText, key []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	blockSize := block.BlockSize()
	padding := blockSize - len(plainText)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	plainText = append(plainText, padText...)

	iv := make([]byte, blockSize)
	_, err = io.ReadFull(rand.Reader, iv)
	if err != nil {
		return "", err
	}

	mode := cipher.NewCBCEncrypter(block, iv)
	encrypted := make([]byte, len(plainText))
	mode.CryptBlocks(encrypted, plainText)

	result := append(iv, encrypted...)

	return base64.StdEncoding.EncodeToString(result), nil
}
func BuildEncrypt(deviceID string) (string, string, error) {
	password := []byte("thisisaverysecretkey1234567890")

	key, salt, err := generateKeyAndSalt(password)
	if err != nil {
		fmt.Println(err)
		return "", "", err
	}
	encryptedID, err := encrypt([]byte(deviceID), key)
	if err != nil {
		fmt.Println(err)
		return "", "", err
	}
	return base64.StdEncoding.EncodeToString(salt), encryptedID, nil
}
func decrypt(cipherText, key []byte) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(string(cipherText))
	if err != nil {
		return "", err
	}

	iv := decoded[:16]
	encrypted := decoded[16:]

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	decrypted := make([]byte, len(encrypted))
	mode.CryptBlocks(decrypted, encrypted)

	unpadding := int(decrypted[len(decrypted)-1])
	return string(decrypted[:len(decrypted)-unpadding]), nil
}
func BuildDecrypt(saltBase64 string, encryptedDeviceID string) string {
	password := []byte("thisisaverysecretkey1234567890")
	salt, err := base64.StdEncoding.DecodeString(saltBase64)
	if err != nil {
		fmt.Println("解码盐值失败:", err)
		return ""
	}
	key := pbkdf2.Key(password, salt, 10000, 32, sha256.New)
	decryptedID, err := decrypt([]byte(encryptedDeviceID), key)
	if err != nil {
		fmt.Println("解密失败:", err)
		return ""
	}
	return decryptedID
}

func checkNpublicExists(npublic string) (bool, string) {
	device, err := btldb.ReadDeviceManagerByNpubKey(npublic)
	if err != nil {
		return false, ""
	}
	return true, device.DeviceID
}
func ProcessDeviceRequest(nonce, nPubKey string) (string, string, error) {

	if nonce == "" || nPubKey == "" {
		return "", "", errors.New("nonce or nPubKey cannot be empty")
	}

	if !VerifyNonce(nonce, nPubKey) {
		return "", "", errors.New("invalid or expired nonce")
	}

	flag, deviceId := checkNpublicExists(nPubKey)
	if !flag {
		var device models.DeviceManager
		deviceID, err := GetDeviceID()
		if err != nil {
			return "", "", err
		}
		device.DeviceID = deviceID
		device.Status = 1
		device.NpubKey = nPubKey
		encryptDeviceID, encodedSalt, err := BuildEncrypt(deviceID)
		if err != nil {
			return "", "", err
		}
		device.EncryptDeviceID = encodedSalt
		err = btldb.CreateDeviceManager(&device)
		if err != nil {
			return "", "", err
		}
		return encryptDeviceID, encodedSalt, nil
	}

	deviceID := deviceId
	encryptDeviceID, encodedSalt, err := BuildEncrypt(deviceID)
	if err != nil {
		return "", "", err
	}

	return encryptDeviceID, encodedSalt, nil
}
