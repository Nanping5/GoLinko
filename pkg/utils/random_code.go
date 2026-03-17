package utils

import (
	"crypto/rand"
	"math/big"
)

// GenerateRandomCode 生成指定位数的随机验证码
func GenerateRandomCode(length int) int {
	if length <= 0 {
		return 0
	}
	// 计算最小值和最大值
	// 例如：length=6 时，min=100000, max=999999
	min := big.NewInt(1)
	for i := 1; i < length; i++ {
		min.Mul(min, big.NewInt(10))
	}
	max := new(big.Int).Mul(min, big.NewInt(10))
	max.Sub(max, big.NewInt(1))

	// 计算范围
	rangeVal := new(big.Int).Sub(max, min)
	rangeVal.Add(rangeVal, big.NewInt(1))

	// 生成 [0, range) 范围内的随机数，然后加上 min
	randomVal, err := rand.Int(rand.Reader, rangeVal)
	if err != nil {
		// 如果 crypto/rand 失败，返回一个默认值
		return int(min.Int64())
	}

	result := new(big.Int).Add(randomVal, min)
	return int(result.Int64())
}
