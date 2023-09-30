package handler

import (
	"stock/common/log"

	"github.com/gin-gonic/gin"
)

// ParseFormMiddleware parse form, such as device
func ParseFormMiddleware(c *gin.Context) {
	if err := c.Request.ParseForm(); err != nil {
		log.Errorf("parse form failed: %+v", err)
	}
	c.Next()
}
