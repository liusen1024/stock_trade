package util

import (
	"fmt"
	"math"
)

// FloatRound 浮点数取整
func FloatRound(f float64, n int) float64 {
	return math.Round(f*math.Pow10(n)) / math.Pow10(n)
}

// IsZero 浮点数判断是否为0
func IsZero(f float64) bool {
	return int64(f*10000) == 0
}

// FloatToPCTStr 转百分数
func FloatToPCTStr(f float64, x int) string {
	return fmt.Sprintf("%.2f%%", FloatRound(f*math.Pow10(x), x))
}

// ChangePctDesc 转换成带+-号的百分比数
func ChangePctDesc(pct float64) string {
	s := fmt.Sprintf("%+.2f%%", pct*100)
	// 去掉0之前的符号
	if s == "+0.00%" || s == "-0.00%" {
		return "0.00%"
	}
	return s
}

// WithSigned 将float转换成带有正负号的string
func WithSigned(f float64) string {
	if f > 0 {
		return fmt.Sprintf("+%0.2f", f)
	}
	return fmt.Sprintf("%0.2f", f)
}

// WithPercent 仅将float64转换成带有百分号
func WithPercent(f float64) string {
	s := fmt.Sprintf("%+.2f%%", f)
	// 去掉0之前的符号
	if s == "+0.00%" || s == "-0.00%" {
		return "0.00%"
	}
	return s
}
