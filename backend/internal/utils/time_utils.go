package utils

import "time"

func GetCurrentTime() string {
	return time.Now().Format(time.RFC3339)
}
