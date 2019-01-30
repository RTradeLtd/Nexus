package delegator

import (
	"errors"
	"net/http"
	"strings"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

var (
	errNoAuth      = errors.New("no authentication provided")
	errInvalidAuth = errors.New("invalid authentication provided")
	errExpiredAuth = errors.New("authentication is expired")
)

func getUserFromJWT(
	r *http.Request,
	keyLookup jwt.Keyfunc,
	timeFunc func() time.Time,
) (user string, err error) {
	// Collect the token from the header.
	bearerString := r.Header.Get("Authorization")

	// Split out the actual token from the header.
	splitToken := strings.Split(bearerString, "Bearer ")
	if len(splitToken) < 2 {
		return "", errNoAuth
	}
	tokenString := splitToken[1]

	// Parse takes the token string and a function for looking up the key.
	token, err := jwt.Parse(tokenString, keyLookup)
	if err != nil {
		return "", errInvalidAuth
	}

	// Verify the claims
	var claims jwt.MapClaims
	var ok bool
	if claims, ok = token.Claims.(jwt.MapClaims); !ok || !token.Valid {
		return "", errInvalidAuth
	}

	// Retrieve ID
	if user, ok = claims["id"].(string); !ok || user == "" {
		return "", errInvalidAuth
	}

	// Check expiry
	if int64(claims["exp"].(float64)) < timeFunc().Unix() {
		return "", errExpiredAuth
	}

	return
}
