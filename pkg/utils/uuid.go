package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"

	"github.com/google/uuid"
)

// GenerateUUID 生成标准的 UUID (36位)
func GenerateUUID() string {
	return uuid.New().String()
}

// GenerateShortID 生成随机数字 ID
func GenerateShortID(length int) string {
	format := fmt.Sprintf("%%0%dd", length)
	max := int64(1)
	for i := 0; i < length; i++ {
		max *= 10
	}
	return fmt.Sprintf(format, generateSecureRandomInt(max))
}

// GenerateUserID 生成 U+6位数字的用户 ID (如 U123456)
func GenerateUserID() string {
	return "U" + GenerateShortID(6)
}

// GenerateGroupID 生成 G+6位数字的群组 ID (如 G123456)
func GenerateGroupID() string {
	return "G" + GenerateShortID(6)
}

// GenerateSessionID 生成 S+8位数字的会话 ID
func GenerateSessionID() string {
	return "S" + GenerateShortID(8)
}

// GenerateMessageID 生成 M+10位数字的消息 ID
func GenerateMessageID() string {
	return "M" + GenerateShortID(10)
}

// GenerateApplyID 生成 A+8位数字的申请 ID
func GenerateApplyID() string {
	return "A" + GenerateShortID(8)
}

func generateSecureRandomInt(max int64) int64 {
	n, err := rand.Int(rand.Reader, big.NewInt(max))
	if err != nil {
		return 0
	}
	return n.Int64()
}
