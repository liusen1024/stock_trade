package util

import (
	"context"

	"github.com/gin-gonic/gin"
)

func RPCContext(c *gin.Context) context.Context {
	return c.Request.Context()
}
