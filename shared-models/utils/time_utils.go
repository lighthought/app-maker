package utils

import "time"

// 获取当前时间-字符串
func GetCurrentTime() string {
	return time.Now().UTC().Format(time.RFC3339)
}

// 获取当前时间
func GetTimeNow() time.Time {
	return time.Now()
}

// 获取n秒后的时间
func GetNSecondLater(n int) time.Time {
	return time.Now().Add(time.Duration(n) * time.Second)
}
