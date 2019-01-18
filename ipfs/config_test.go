package ipfs

import (
	"fmt"
	"io/ioutil"
	"reflect"
	"testing"
)

func Test_getConfig(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		want    *GoIPFSConfig
		wantErr bool
	}{
		{"no file", args{"./what"}, nil, true},
		{"wrong file", args{"./ipfs.go"}, nil, true},
		{"right file",
			args{"../testenv/ipfs/nodes/1/config"},
			&GoIPFSConfig{struct{ PrivKey string }{"CAASpgkwggSiAgEAAoIBAQCqbaLjmbfDoABV3FxXlYkiiAWzQee2k+kFBRgKiDdGORH9eaH951ifv2vHe3xKUROMNaGTRvN4sIeI7tBSCrBSDRJ7PgS5zcdLwKPl8DID7/kkZ+QK7BmEnIcFvCCFxOc3DUROu7oT16kH4chWqoEWPH0SWp3DjIN3zN6FC0oizCKTZeg2HZDZX1nkIhpkKPs3Bp0X4kQVA6EW5xugez3uzbkVcxGzBvPtArisXlTE/GLG31SwrF8OnYbAqZB+TtEm5i7r5kHj4iJScsv1xBVxrdzBDOE16KW3inDGxPnW6vwOPzcN4fTMZ3wgRk89Cm5iI/MmtkiaUjGJPr08+xWzAgMBAAECggEAbla5BN36qX6neO9IIbRAqsih2CKtH/m2/XcEz5zNHHvKd+8Nv9LN/+7wmqAKIhtHqpj2WOGws8ymkzL6UIN3EEhCVOQcLydZBmRcOHxABWiSRs20SJX/F2o3yLC55aFLiMrgFJFZsYsIdn/pMqMFHB5hY0ajqX0JiMBsuHpMryWnm/5sllp7MFc6yEQF0UQ+z+Dx83hYoh/40qwHiuohm7i1MQYx9qKjMMKLMoHdGlg6g3nMRgeXw1xI8ry50A2AQjcQsbpYN+/FxASnf9l1VuIFgOnzGmbBGGC5I+SStWoupinrNH68MuQxCqmIiexYPCFvPW/YjThIMqYZ37+ooQKBgQDDxDdJ6G899qqbrR4ly/XWbe+CDxf8ueMvWsirvkL3ozufsvNPEIrXXKxuMex0JB5x7wPPr7jhwSNVuGD1bsu8R2b66n985rbTgFT1B5eGuW9FHuhPoC+bNeawi/W7ZsCQKCsUHY1orA2NAanaSnzSS9ico3F6gQ/txafwoIMlCwKBgQDe3aBC4TFZa79nRpOyPiTU8y5iPSJ8xf/6f2GA9voXLLnII8RLZXVtc71zj1OIDks//fNwDnqdwcczalqCfNSdrXWavRJ7dqKOHPQBpO8eGNHtYBd876ushbUR4gjrkxgXZ2uyrXe+D8TcDHWCdCJ3aa3FjzuMv/JJ4Fe/ErDq+QKBgDRdNdTFIYxXgIcnpVrC1b1HprsJQodNSaGPDQIzYEJRHU+4VDCf4iN9HHpVTEQ8rRAYuNJC1Jc+TC9PpE/CFSkFiFwxgWxtYhXsy8zG/RcCXusEO2uhE1rW7h/nMBGyiGuG8w7sYLjQ3McM3NwQ9JZjx0sOxPnZr+MP7b4FkU7FAoGASFS5rLsVnyX/Ku+XA+RzY8HBLhUVWlWQrKYm6Qo/RMI5UaF6FdZJ9En6FMVRoPiyp4QuPBIW7Zh0pFVCJtOI1dv0LVJr6zIns+PltZroGGaJy3bCaMQIfaeviqxHpN1Kll30cDsof8DybVCF2t8CSKs9wL6p3xZ09lEfaV4RmVECgYASBPs0qAdJ+WS2i8OErkN/HZRfNQwRo83YPeYTbhhe1oHfXyQqBM1CC965+Xbh82t8xYPG3Nc1Av63Rfh/D9+J3ylduUqpNxN0xgdQWYXzDA7Lzz+GkHwGGVAlbabK966m6PDZIpCxgRU/pLbUQ9YIJUORCPXp6psl/yWTgn9mgA=="}},
			false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getConfig(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("getConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_newNodeStartScript(t *testing.T) {
	var (
		disk = 10
	)

	f, _ := ioutil.ReadFile("./internal/ipfs_start.sh")
	expected := fmt.Sprintf(string(f), disk)

	got, err := newNodeStartScript(disk)
	if err != nil {
		t.Error("unexpected err:", err.Error())
		return
	}
	if got != expected {
		t.Errorf("expected '%s', got '%s'", expected, got)
	}
}
