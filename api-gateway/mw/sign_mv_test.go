package mw

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"stock/api-gateway/util"
	"stock/common/env"

	"github.com/gin-gonic/gin"
)

func TestAESEncrypt(t *testing.T) {
	tm := time.Now()
	r := RandomSecret(tm)
	fmt.Printf("key = %v\n", r)
	key := []byte(fmt.Sprintf("%16b", r))

	rd1 := "sdj989798ert2-123"
	rd, _ := util.AESEncrypt(rd1, key)
	fmt.Printf("rd= %s\n", rd)

	crd, err := decryptRandomString(rd, tm)
	if err != nil {
		t.Fatal(err)
	}
	if crd != rd1 {
		t.Fatal()
	}
	fmt.Printf("rd= %s\n", crd)

	// sign
	values := url.Values{
		"a":  []string{"1"},
		"b":  []string{"2"},
		"rd": []string{rd},
		"sn": []string{"u8Ha6wEwE3V0o8ta5zXNuw=="}, // key=5669
	}
	sn := Sign(values, rd1)
	fmt.Printf("sn=%s\n", sn)
}

func TestVerifySign(t *testing.T) {
	env.LoadGlobalEnv("../conf/test.json")
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = &http.Request{}
	c.Request.Form = url.Values{
		"a":  []string{"1"},
		"b":  []string{"2"},
		"rd": []string{"GFENZu2Ne9wOhrv5IURwEAkAosXfSr3nswNJ-2VjFmv9sEKJhDPn6oO62En9g5xe"},
		"sn": []string{"39u_gC5qRUU7oI9oQu5KeQ="}, // key=5669
	}
	VerifySignMiddleware(c)
	fmt.Printf("%v\n", c.Request.Form)
}
