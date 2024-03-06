package number

import "math"

// Lua语言层面一共有25个运算符，按类别可以分为算术（Arithmetic）运算符、按位（Bitwise）运算符、比较（Comparison）运算符、逻辑（Logical）运算符、长度运算符和字符串拼接运算符

// 算术运算符共8个，分别是：+（加）、-（减、一元取反）、*（乘）、/（除）、//（整除）、%（取模）、^（乘方

func FloatToInteger(f float64) (int64, bool) {
	i := int64(f)
	return i, float64(i) == f
}

// a % b == a - ((a // b) * b)
func IMod(a, b int64) int64 {
	return a - IFloorDiv(a, b)*b
}

// a % b == a - ((a // b) * b)
func FMod(a, b float64) float64 {
	if a > 0 && math.IsInf(b, 1) || a < 0 && math.IsInf(b, -1) {
		return a
	}
	if a > 0 && math.IsInf(b, -1) || a < 0 && math.IsInf(b, 1) {
		return b
	}
	return a - math.Floor(a/b)*b
}

func IFloorDiv(a, b int64) int64 {
	if a > 0 && b > 0 || a < 0 && b < 0 || a%b == 0 {
		return a / b
	} else {
		return a/b - 1
	}
}

func FFloorDiv(a, b float64) float64 {
	return math.Floor(a / b)
}

func ShiftLeft(a, n int64) int64 {
	if n >= 0 {
		return a << uint64(n)
	} else {
		return ShiftRight(a, -n)
	}
}

func ShiftRight(a, n int64) int64 {
	if n >= 0 {
		return int64(uint64(a) >> uint64(n))
	} else {
		return ShiftLeft(a, -n)
	}
}