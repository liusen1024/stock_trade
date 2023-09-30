package util

import (
	"time"
)

func TimeToInt32(v time.Time) int32 {
	return int32(v.Year()*1e4 + int(v.Month())*1e2 + v.Day())
}

func TimeToInt64(v time.Time) int64 {
	return int64(v.Year()*1e13 + int(v.Month())*1e11 + v.Day()*1e9 +
		v.Hour()*1e7 + v.Minute()*1e5 + v.Second()*1e3 + v.Nanosecond()/1e6)
}

func Int32ToTime(v int32) time.Time {
	y := int(v) / 1e4
	m := (int(v) - y*1e4) / 1e2
	d := int(v) - y*1e4 - m*1e2
	return time.Date(y, time.Month(m), d, 0, 0, 0, 0, time.Local)
}

func Int64ToTime(v int64) time.Time {
	y := int(v) / 1e13
	m := (int(v) - y*1e13) / 1e11
	d := (int(v) - y*1e13 - m*1e11) / 1e9
	h := (int(v) - y*1e13 - m*1e11 - d*1e9) / 1e7
	M := (int(v) - y*1e13 - m*1e11 - d*1e9 - h*1e7) / 1e5
	s := (int(v) - y*1e13 - m*1e11 - d*1e9 - h*1e7 - M*1e5) / 1e3
	sss := (int(v) - y*1e13 - m*1e11 - d*1e9 - h*1e7 - M*1e5 - s*1e3) * 1e6

	return time.Date(y, time.Month(m), d, h, M, s, sss, time.Local)
}
