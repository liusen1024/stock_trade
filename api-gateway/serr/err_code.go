package serr

const (
	// ErrCodeInvalidParam 参数错误
	ErrCodeInvalidParam = "1"
	// ErrCodeBusinessFail 业务失败
	ErrCodeBusinessFail = "400"
	// ErrCodeContractNoFound 合约不存在
	ErrCodeContractNoFound = "401"
	// SuccessCode 业务成功状态码
	SuccessCode = "200"
	// ErrCodeNoLogin 未登录
	ErrCodeNoLogin = "300"
	// ErrCodeDataNoFound 数据不存在
	ErrCodeDataNoFound = "2"
)

// ErrUserNotLogin 用户未登录
//var ErrUserNotLogin = hxerr.New(ErrCodeSessionExpired, "用户未登录")
