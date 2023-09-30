package handler

import (
	"errors"
	"net/http"
	"stock/api-gateway/serr"

	"github.com/gin-gonic/gin"
)

// JSONWrapper 将数据处理函数封装为json接口返回
func JSONWrapper(fn func(*gin.Context) (interface{}, error)) func(*gin.Context) {
	return func(c *gin.Context) {
		data, err := fn(c)
		// TODO 这里可以做一些具体错误的上报和封装，不一定要将开发消息发送给客户端，
		// 可以根据error code，转化为更人性化的错误消息
		if err != nil {
			var StockError *serr.StockError
			if errors.As(err, &StockError) {
				c.JSON(http.StatusOK, map[string]interface{}{
					"code":    "200",
					"message": StockError.Msg,
					"status":  "fail",
				})
				return
			}

			c.JSON(http.StatusOK, map[string]interface{}{
				"code":    "200",
				"message": err.Error(),
				"status":  "fail",
			})
			return
		}

		c.JSON(http.StatusOK, data)
	}
}

// PureJSONWrapper 将数据处理函数封装为json接口返回
func PureJSONWrapper(fn func(*gin.Context) (interface{}, error)) func(*gin.Context) {
	return func(c *gin.Context) {
		data, err := fn(c)
		// TODO 这里可以做一些具体错误的上报和封装，不一定要将开发消息发送给客户端，
		// 可以根据error code，转化为更人性化的错误消息
		if err != nil {
			var StockError *serr.StockError
			if errors.As(err, &StockError) {
				c.PureJSON(http.StatusOK, map[string]interface{}{
					"code":    "200",
					"message": StockError.Msg,
					"status":  "fail",
				})
				return
			}

			c.PureJSON(http.StatusOK, map[string]interface{}{
				"code":    "200",
				"message": err.Error(),
				"status":  "fail",
			})
			return
		}

		c.PureJSON(http.StatusOK, data)
	}
}
