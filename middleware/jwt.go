package middleware

import (
	"errors"
	"github.com/dgrijalva/jwt-go"
	"github.com/go-redis/redis/v8"
	"time"
	"trade/config"
)

var (
	jwtKey = []byte("my_secret_key")
)

type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

func GenerateToken(username string) (string, error) {
	expirationTimeMinute := config.GetLoadConfig().Redis.ExpirationTimeMinute
	if expirationTimeMinute == 0 {
		expirationTimeMinute = 100 * 365 * 24 * 60
	}
	expirationTime := time.Now().Add(time.Duration(expirationTimeMinute) * time.Minute)
	claims := &Claims{
		Username: username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return "", err
	}
	validateToken, err := RedisGet(username)
	if validateToken != "" {

		return validateToken, nil
	}

	if err != nil && !errors.Is(err, redis.Nil) {
		return "", err
	}

	redisSetTimeMinute := config.GetLoadConfig().Redis.RedisSetTimeMinute
	err = RedisSet(username, tokenString, time.Duration(redisSetTimeMinute)*time.Minute)
	if err != nil {
		return "", err
	}
	err = RedisSet(tokenString, username, time.Duration(redisSetTimeMinute)*time.Minute)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func ValidateToken(tokenString string) (*Claims, error) {

	_, err := RedisGet(tokenString)
	if err != nil {
		return nil, errors.New("invalid token")
	}
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}
