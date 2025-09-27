package utils

import "time"

// NowUTC 返回当前 UTC 时间字符串
func NowUTC() string {
	return time.Now().UTC().Format(time.RFC3339)
}
