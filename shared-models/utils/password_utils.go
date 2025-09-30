package utils

import (
	"crypto/rand"
	"math/big"
)

// PasswordUtils 密码工具类
type PasswordUtils struct{}

// NewPasswordUtils 创建密码工具实例
func NewPasswordUtils() *PasswordUtils {
	return &PasswordUtils{}
}

// GenerateRandomPassword 生成随机密码
// 必须包含：大写字母、小写字母、数字、特殊符号
func (u *PasswordUtils) GenerateRandomPassword(prefix string) string {
	const (
		lowercase = "abcdefghijklmnopqrstuvwxyz"
		uppercase = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
		digits    = "0123456789"
		symbols   = "!@#$%^&*()_+-=[]{}|;:,.<>?"
		allChars  = lowercase + uppercase + digits + symbols
	)

	// 确保每种字符类型至少有一个
	password := make([]byte, 0, 16)

	// 添加前缀（如果提供）
	if prefix != "" {
		password = append(password, []byte(prefix)...)
	}

	// 确保包含每种字符类型
	password = append(password, u.randomChar(lowercase)) // 小写字母
	password = append(password, u.randomChar(uppercase)) // 大写字母
	password = append(password, u.randomChar(digits))    // 数字
	password = append(password, u.randomChar(symbols))   // 特殊符号

	// 填充剩余长度到16位
	remainingLength := 16 - len(password)
	for i := 0; i < remainingLength; i++ {
		password = append(password, u.randomChar(allChars))
	}

	// 打乱密码字符顺序（除了前缀）
	if prefix != "" {
		prefixLen := len(prefix)
		shuffledPart := password[prefixLen:]
		u.shuffleBytes(shuffledPart)
		copy(password[prefixLen:], shuffledPart)
	} else {
		u.shuffleBytes(password)
	}

	return string(password)
}

// randomChar 从指定字符集中随机选择一个字符
func (u *PasswordUtils) randomChar(chars string) byte {
	n, err := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
	if err != nil {
		// 如果加密随机数生成失败，回退到UUID
		return chars[0]
	}
	return chars[n.Int64()]
}

// shuffleBytes 打乱字节数组顺序
func (u *PasswordUtils) shuffleBytes(bytes []byte) {
	for i := len(bytes) - 1; i > 0; i-- {
		j, err := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		if err != nil {
			continue
		}
		bytes[i], bytes[j.Int64()] = bytes[j.Int64()], bytes[i]
	}
}

// GenerateSimplePassword 生成简单密码（仅用于测试）
func (u *PasswordUtils) GenerateSimplePassword() string {
	n := GenerateUUID()
	return n[:12]
}
