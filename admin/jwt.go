package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
)

func createJWTtoken(login string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.MapClaims{
		"login": login,
		"exp":   time.Now().AddDate(0, 1, 0).Unix(),
	})

	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_KEY")))
	if err != nil {
		return "", fmt.Errorf("cannot get signed token string: %v", err)
	}

	return tokenString, nil
}

func getJWTtokenFromCookies(cookies []*http.Cookie) (*jwt.Token, error) {
	var jwtCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "jwt" {
			jwtCookie = c
		}
	}
	if jwtCookie == nil {
		return nil, fmt.Errorf("jwt token was not found in cookies.")
	}

	return parseJWTtoken(jwtCookie.Value)
}

func parseJWTtoken(token string) (*jwt.Token, error) {
	return jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		jwtKey := os.Getenv("JWT_KEY")
		if jwtKey == "" {
			return nil, fmt.Errorf("JWT_KEY was not found in env")
		}
		return []byte(jwtKey), nil
	})
}
