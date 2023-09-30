package serr

import (
	"fmt"
)

type StockError struct {
	Code string // 具体错误
	Msg  string // 可以反馈给用户看的消息
}

func (err *StockError) Error() string {
	return fmt.Sprintf("code: [%s], msg: [%s]", err.Code, err.Msg)
}

func New(code string, msg string) *StockError {
	return &StockError{
		Code: code,
		Msg:  msg,
	}
}

func Errorf(code string, format string, a ...interface{}) *StockError {
	return &StockError{
		Code: code,
		Msg:  fmt.Sprintf(format, a...),
	}
}

// ErrBusiness 业务错误
func ErrBusiness(msg string) *StockError {
	return New(ErrCodeBusinessFail, msg)
}

// ErrNoLogin 未登录错误
func ErrNoLogin() *StockError {
	return New(ErrCodeNoLogin, "请登录")
}
