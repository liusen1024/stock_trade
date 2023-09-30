package util

import (
	"crypto/md5"
	"fmt"
	"io"
)

// MD5 计算 md5
func MD5(str string) string {
	w := md5.New()
	io.WriteString(w, str)
	md5str := fmt.Sprintf("%x", w.Sum(nil))
	return md5str
}
