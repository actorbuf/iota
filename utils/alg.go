package utils

import "math"

// DivisionRoundingUp 除法a/b向上取整
func DivisionRoundingUp(a, b int64) int64 {
	fa := float64(a)
	fb := float64(b)
	return int64(math.Ceil(fa / fb))
}

// DivisionRoundingDown 除法a/b向下取整
func DivisionRoundingDown(a, b int64) int64 {
	fa := float64(a)
	fb := float64(b)
	return int64(math.Floor(fa / fb))
}

// DivisionRounding 除法a/b四舍五入
func DivisionRounding(a, b int64) int64 {
	fa := float64(a)
	fb := float64(b)
	return int64(math.Floor((fa / fb) + 0.5))
}
