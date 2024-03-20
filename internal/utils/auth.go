package utils

import (
	"strings"

	"github.com/dgrijalva/jwt-go"
)

var secretKey = []byte("secretkey")

// GenerateToken генерирует JWT-токен на основе переданного имени пользователя.
func GenerateToken() (string, error) {

	t := jwt.New(jwt.SigningMethodHS256)
	s, er := t.SignedString(secretKey)
	return s, er
}

func VerifyToken(tokenString string) (jwt.MapClaims, error) {

	// Проверка и удаление префикса "Bearer "
	if strings.HasPrefix(tokenString, "Bearer ") {
		tokenString = strings.TrimPrefix(tokenString, "Bearer ")
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil

	})

	if err != nil || !token.Valid {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, jwt.ErrInvalidKeyType
	}

	return claims, nil
}
