package errgroup

import (
	"stock/common/log"

	"golang.org/x/sync/errgroup"
)

type Group struct {
	errgroup.Group
	cnt int
	ch  chan bool
}

// GroupWithCount 控制总协程数
func GroupWithCount(cnt int) *Group {
	var g Group
	if cnt > 0 {
		g.cnt = cnt
		g.ch = make(chan bool, cnt)
	}
	return &g
}

func (g *Group) Go(fn func() error) {
	if g.cnt > 0 {
		g.ch <- true
	}
	wrapper := func() error {
		defer func() {
			if r := recover(); r != nil {
				log.Errorf("recover %v", r)
			}
			if g.cnt > 0 {
				<-g.ch
			}
		}()
		return fn()
	}
	g.Group.Go(wrapper)
}
