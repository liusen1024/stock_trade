package handler

import (
	"net/http"
	"stock/api-gateway/serr"
	"stock/common/log"
	"time"

	"github.com/gin-gonic/gin"
)

var excludeMap = map[string]bool{
	"/cms/login":       true,
	"/cms/source_path": true,
	"/cms/logout":      true,
}

// Auth 检查是否登录
func Auth(c *gin.Context) {
	if excludeMap[c.Request.URL.Path] {
		c.Next()
		return
	}
	ctx := c.Request.Context()
	cookie, err := c.Cookie("x-token")
	if err != nil {
		c.JSON(http.StatusOK, map[string]interface{}{
			"code": serr.ErrCodeNoLogin,
			"msg":  "鉴权无效",
		})
		c.Abort()
		return
	}
	session, err := GetSession(ctx, cookie)
	if err != nil || session == nil {
		// 客户端需要重新获取短票
		c.JSON(http.StatusOK, map[string]interface{}{
			"code": serr.ErrCodeNoLogin,
			"msg":  "鉴权无效",
		})
		c.Abort()
		return
	}

	c.Set("__USERNAME", session.UserName)
	log.Infof("request URL: %s, username: %s, req time: %s", c.Request.RequestURI, session.UserName, time.Now().Format("2006-01-02 15:04:05"))
	c.Next()
}

// ACL 检查是否登录
func ACL(c *gin.Context) {
	if excludeMap[c.Request.URL.Path] {
		c.Next()
		return
	}
	if c.Request.Method != "GET" && Username(c) != "admin" {
		c.JSON(http.StatusOK, map[string]interface{}{
			"code": serr.ErrCodeBusinessFail,
			"msg":  "权限不足",
		})
		c.Abort()
		return
	}

	c.Next()
}

// ParseFormMiddleware parse form, such as device
func ParseFormMiddleware(c *gin.Context) {
	if err := c.Request.ParseForm(); err != nil {
		log.Errorf("parse form failed: %+v", err)
	}
	c.Next()
}
