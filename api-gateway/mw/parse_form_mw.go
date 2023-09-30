package mw

import (
	"bytes"
	"io/ioutil"

	"stock/common/log"

	"github.com/gin-gonic/gin"
)

// ParseFormMiddleware parse form, such as device
func ParseFormMiddleware(c *gin.Context) {
	if err := c.Request.ParseForm(); err != nil {
		log.Errorf("parse form failed: %v", err)
	}

	// 在middleware中调用一次bind读完后，后续无法使用body
	// gin里面的ShouldBindBodyWith也不是很好
	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		log.Errorf("read request body error: %v", err)
	}
	c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(body))

	//// 绑定设备
	//var d model.Device
	//if err := c.ShouldBind(&d); err != nil {
	//	log.Errorf("bind device failed: %v", err)
	//}
	//c.Set("__device__", &d)

	// 后续接着使用body
	c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(body))

	c.Next()
}
