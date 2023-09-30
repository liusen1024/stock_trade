package mw

import (
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"stock/api-gateway/util"
	"stock/common/env"
	"stock/common/log"

	"github.com/gin-gonic/gin"
)

// VerifySignMiddleware 签名验证
func VerifySignMiddleware(c *gin.Context) {
	// 非生产环境不验签
	if !env.GlobalEnv().IsProd() {
		c.Next()
		return
	}

	// 开始验证签名
	// 获取随机串
	rd := c.Request.Form.Get("rd")
	if len(rd) == 0 {
		c.JSON(http.StatusOK, map[string]interface{}{
			"code": -1,
			"msg":  "no rd param",
		})
		return
	}

	// 使用这个随机串作为盐，去md5所有参数，得到签名
	sn := c.Request.Form.Get("sn")
	if len(sn) == 0 {
		c.JSON(http.StatusOK, map[string]interface{}{
			"code": -1,
			"msg":  "no sn param",
		})
		return
	}

	fn := func(tm time.Time) (bool, error) {
		rd, err := decryptRandomString(rd, tm)
		if err != nil {
			log.Warnf("rd invalid %s", rd)
			return false, err
		}
		expect := Sign(c.Request.Form, rd)
		if sn != expect {
			log.Warnf("sign invalid %s != %s(expect)", sn, expect)
			return false, nil
		}
		return true, nil
	}

	// 使用今天或者昨天的日期当作密钥来解这个随机串
	ok, err := fn(time.Now())
	if err != nil {
		c.JSON(http.StatusOK, map[string]interface{}{
			"code": -1,
			"msg":  err.Error(),
		})
		return
	}
	if !ok {
		ok, err = fn(time.Now().AddDate(0, 0, -1))
	}
	if !ok || err != nil {
		c.JSON(http.StatusOK, map[string]interface{}{
			"code": -1,
			"msg":  "invalid sn param",
		})
		return
	}

	// 签名正常
	c.Next()
}

// Sign return signature
func Sign(values url.Values, salt string) string {
	list := make([]string, 0, len(values))
	for k, v := range values {
		if k == "sn" {
			continue
		}
		kv := k + "="
		if len(v) > 0 {
			kv += v[0]
		}
		list = append(list, kv)
	}
	sort.Strings(list)
	buf := strings.Join(list, "|")
	buf += salt

	h := md5.New()
	if _, err := io.WriteString(h, buf); err != nil {
		return ""
	}
	return base64.URLEncoding.EncodeToString(h.Sum(nil))
}

func decryptRandomString(ciphertext string, tm time.Time) (string, error) {
	key := RandomSecret(tm)
	secret := []byte(fmt.Sprintf("%16b", key))
	return util.AESDecrypt(ciphertext, secret)
}
