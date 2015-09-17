package faas

import (
	//  "log"
	// "net/http"
	//  "bytes"
	//  "encoding/gob"
	"time"
	//  "io/ioutil"
	//  "encoding/json"
	"fmt"
	"github.com/dgrijalva/jwt-go"
)

/*
type AuthToken struct {
  Email string `json:"email"`
  Created time.Time  `json:"created"`
  Updated time.Time  `json:"updated"`
}
*/

const (
	SecretKey       = "21371682371623879451235182"
	TokenExpMinutes = 60
)

func GetAuthToken(user *userInfo) (string, error) {
	token := jwt.New(jwt.GetSigningMethod("HS256"))
	//token.Claims["id"] = user.Id
	token.Claims["email"] = user.Email
	// Expire in 5 mins
	token.Claims["exp"] = time.Now().Add(time.Minute * TokenExpMinutes).Unix()
	tokenString, err := token.SignedString([]byte(SecretKey))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func UpdateAuthToken(authToken string) (string, error) {

	oldToken, err := jwt.Parse(authToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(SecretKey), nil
	})

	if err == nil && oldToken.Valid {

		claims := oldToken.Claims

		newToken := jwt.New(jwt.GetSigningMethod("HS256"))

		//newToken.Claims["id"] = claims["id"].(string)
		newToken.Claims["email"] = claims["email"].(string)
		// Expire in 5 mins
		newToken.Claims["exp"] = time.Now().Add(time.Minute * TokenExpMinutes).Unix()
		tokenString, err := newToken.SignedString([]byte(SecretKey))
		if err != nil {
			return "", err
		}
		return tokenString, nil
	}

	return "", err

}

func CheckAuthToken(authToken string) (string, error) {
	token, err := jwt.Parse(authToken, func(token *jwt.Token) (interface{}, error) {

		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(SecretKey), nil
	})

	if err == nil && token.Valid {
		claims := token.Claims
		//log.Printf("%s,%s",claims["id"],claims["email"])
		return /*claims["id"].(string), */claims["email"].(string), nil
	} else {
		return "", err
	}
}
