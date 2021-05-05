package http2

import (
	"testing"
	"time"
)

func TestHttpGet(t *testing.T) {
	url := "https://www.baidu.com"
	b, err := GetHttp(url, 5*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf(string(b))
}
