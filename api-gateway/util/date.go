package util

import "time"

// PreciseFormat 精准格式化，格式化到秒
func PreciseFormat(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

// FuzzyFormat 模糊格式化，格式化到天
func FuzzyFormat(t time.Time) string {
	return t.Format("2006-01-02")
}

// Bod BeginningOfDay
func Bod(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, t.Location())
}
