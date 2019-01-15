package mock

import jwt "github.com/dgrijalva/jwt-go"

// FakeToken generates a fake JWT token
func FakeToken(user, key string) string {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["id"] = user
	tokenString, _ := token.SignedString([]byte(key))
	return tokenString
}
