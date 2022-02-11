package core

import "fmt"

const UnknownError = "unknown"

const (
	// ErrNil 正常
	ErrNil = 0
	// ErrSystemError 系统错误
	ErrSystemError = -1
	// ErrProcessPanic 服务PANIC
	ErrProcessPanic = -10
)

const (
	// ErrInvalidArg 请求参数无效
	ErrInvalidArg = 1001
	// ErrRecordNotFound 找不到记录
	ErrRecordNotFound = 1002
	// ErrConnectTimeout 连接超时
	ErrConnectTimeout = 1003
	// ErrFreqLimit 频率限制 [业务级]
	ErrFreqLimit = 1004
	// ErrRequestBroken 请求熔断
	ErrRequestBroken = 1005
	// ErrRequestRateLimit 请求限流 [服务级]
	ErrRequestRateLimit = 1006
	// ErrParamEmpty 请求参数为空
	ErrParamEmpty = 1007
)

var (
	errCodeMap = map[int32]string{
		ErrNil:              "ok",
		ErrSystemError:      "system error",
		ErrProcessPanic:     "process panic",
		ErrInvalidArg:       "invalid arg",
		ErrRecordNotFound:   "record not found",
		ErrConnectTimeout:   "connect timeout",
		ErrFreqLimit:        "request freq limit",
		ErrRequestBroken:    "request is broken",
		ErrRequestRateLimit: "request rate is limited",
		ErrParamEmpty:       "request param is empty",
	}
)

func (m *ErrMsg) Error() string {
	return fmt.Sprintf("err_code: %d, err_msg: %s", m.ErrCode, m.ErrMsg)
}

func RegisterError(m map[int32]string) {
	for k, v := range m {
		errCodeMap[k] = v
	}
}

// GetErrMsg 基于错误码返回错误信息
func GetErrMsg(errCode int32) string {
	if errCode == 0 {
		return "success"
	}

	msg, ok := errCodeMap[errCode]
	if ok {
		return msg
	}
	if errCode < 0 {
		return "system error"
	}

	return UnknownError
}

// CreateError 基于错误码返回错误信息
func CreateError(errCode int32) *ErrMsg {
	return &ErrMsg{ErrCode: errCode, ErrMsg: errCodeMap[errCode]}
}

// GetErrCode 获取错误码 默认1 业务错误
func GetErrCode(err error) int {
	if err == nil {
		return 0
	}

	if p, ok := err.(*ErrMsg); ok {
		return int(p.ErrCode)
	}

	return 1
}

// CreateErrorWithMsg 自定义创建错误 如果错误码已提前注册 errMsg将失效
func CreateErrorWithMsg(errCode int32, errMsg string) *ErrMsg {
	return &ErrMsg{ErrCode: errCode, ErrMsg: errMsg, Autonomy: true}
}
