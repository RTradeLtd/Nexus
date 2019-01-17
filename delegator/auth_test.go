package delegator

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	jwt "github.com/dgrijalva/jwt-go"
)

var (
	defaultTestKey = []byte("suchStrongKeyMuchProtectVerySafe")

	// validToken is a token generated using Temporal's gin-jwt configuration,
	// signed using defaultTestKey (the same used in RTradeLtd/testenv), generated
	// for user 'testuser'
	validToken = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1NDc3Nzg2MzcsImlkIjoidGVzdHVzZXIiLCJvcmlnX2lhdCI6MTU0NzY5MjIzN30.2oqQCym2mcyFl8mjOHoGNtK41SJLwX0xbWScDruDECQ"
)

func getTestKey(*jwt.Token) (interface{}, error) { return defaultTestKey, nil }

func Test_getUserFromJWT(t *testing.T) {
	type args struct {
		r         *http.Request
		keyLookup jwt.Keyfunc
	}
	tests := []struct {
		name     string
		args     args
		wantUser string
		wantErr  bool
	}{
		{"no header",
			args{httptest.NewRequest("", "/", nil), getTestKey},
			"", true},
		{"invalid header",
			args{func() *http.Request {
				var r = httptest.NewRequest("", "/", nil)
				r.Header.Set("Authorization", "asdf")
				return r
			}(), getTestKey},
			"", true},
		{"invalid token",
			args{func() *http.Request {
				var r = httptest.NewRequest("", "/", nil)
				r.Header.Set("Authorization", "Bearer asdf")
				return r
			}(), getTestKey},
			"", true},
		{"valid token, should return correct user",
			args{func() *http.Request {
				var r = httptest.NewRequest("", "/", nil)
				r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", validToken))
				return r
			}(), getTestKey},
			"testuser", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotUser, err := getUserFromJWT(tt.args.r, tt.args.keyLookup)
			if (err != nil) != tt.wantErr {
				t.Errorf("getUserFromJWT() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotUser != tt.wantUser {
				t.Errorf("getUserFromJWT() = %v, want %v", gotUser, tt.wantUser)
			}
		})
	}
}
