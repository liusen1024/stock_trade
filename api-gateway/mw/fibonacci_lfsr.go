package mw

import (
	"time"
)

// FibonacciLFSR 线性反馈移位寄存器
type FibonacciLFSR struct {
	state int
}

func (f *FibonacciLFSR) Next() int {
	//  影响下一个状态的比特位叫做抽头。图中，抽头序列为[16,14,13,11]。LFSR最右端的比特为输出比特。抽头依次与输出比特进行异或运算，然后反馈回最左端的位。最右端位置所生成的序列被称为输出流。
	// https://www.knowpia.cn/pages/%E7%BA%BF%E6%80%A7%E5%8F%8D%E9%A6%88%E7%A7%BB%E4%BD%8D%E5%AF%84%E5%AD%98%E5%99%A8
	// 假设抽头是 16,14,13,11
	a := f.state & 1        // 取16位
	b := (f.state >> 2) & 1 // 取14位
	c := (f.state >> 3) & 1 // 取13位
	d := (f.state >> 5) & 1 // 取11位
	out := a ^ b ^ c ^ d
	f.state = f.state>>1 | (out << 15)
	return f.state
}

const start = 12984

var startDate = time.Date(1921, 7, 23, 0, 0, 0, 0, time.Local)

func RandomSecret(tm time.Time) int {
	y, m, d := tm.Date()
	tm = time.Date(y, m, d, 0, 0, 0, 0, time.Local)
	f := &FibonacciLFSR{
		state: start,
	}
	for startDate.Before(tm) || startDate.Equal(tm) {
		tm = tm.AddDate(0, 0, -1)
		f.Next()
	}
	return f.state
}
