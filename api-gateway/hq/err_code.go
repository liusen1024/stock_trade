package handler

import "fmt"

type HqErr struct {
	Code string // 具体错误
	Msg  string // 可以反馈给用户看的消息
}

func (err *HqErr) Error() string {
	return fmt.Sprintf("code: [%s], msg: [%s]", err.Code, err.Msg)
}

func New(code string, msg string) *HqErr {
	return &HqErr{
		Code: code,
		Msg:  msg,
	}
}

func Errorf(code string, format string, a ...interface{}) *HqErr {
	return &HqErr{
		Code: code,
		Msg:  fmt.Sprintf(format, a...),
	}
}
