package util

import (
	"context"
	"github.com/namejlt/gozen"
	"time"
)

type LimitCh struct {
	ch       chan bool
	limitNum int
	ticker   time.Duration
	ctx      context.Context
	cancel   context.CancelFunc
}

func NewLimitCh(limitNum int, ticker time.Duration) (p *LimitCh) {
	p = new(LimitCh)
	p.limitNum = limitNum
	p.ch = make(chan bool, limitNum)
	p.ticker = ticker
	p.ctx, p.cancel = context.WithCancel(context.TODO())
	return
}

func (p *LimitCh) Start() {
	p.initData()
	t := time.NewTicker(p.ticker)
	go gozen.GoFuncOne(func() error {
		for {
			select {
			case <-p.ctx.Done():
				close(p.ch)
				return nil
			case <-t.C:
				p.initData()
			}
		}
	})
}

func (p *LimitCh) initData() {
	i := len(p.ch)
	for ; i < p.limitNum; i++ {
		p.ch <- true
	}
}

func (p *LimitCh) Consumer() bool {
	return <-p.ch
}

func (p *LimitCh) Close() {
	p.cancel()
	return
}
