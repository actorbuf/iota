package utils

import "time"

// GetTimeNowUnix 获取当前时间的标准时间戳
func GetTimeNowUnix() int64 {
	return time.Now().Unix()
}

// GetTimeNowUnixMilli 获取当前时间的毫秒时间戳
func GetTimeNowUnixMilli() int64 {
	return time.Now().UnixNano() / 1e6
}

// UnixMilli2Time 毫秒时间戳转 time.Time 结构
func UnixMilli2Time(u int64) time.Time {
	return time.Unix(u/1e3, 0)
}
