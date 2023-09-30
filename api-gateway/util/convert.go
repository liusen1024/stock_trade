package util

import (
	"strconv"

	"github.com/axgle/mahonia"
)

func String2Float64(s string) float64 {
	value, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return value
}

func String2Int64(s string) int64 {
	value, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0
	}
	return value
}

func ConvertToString(src string, srcCode, tagCode string) string {
	srcCoder := mahonia.NewDecoder(srcCode)

	srcResult := srcCoder.ConvertString(src)

	tagCoder := mahonia.NewDecoder(tagCode)

	_, cdata, _ := tagCoder.Translate([]byte(srcResult), true)

	result := string(cdata)

	return result

}
