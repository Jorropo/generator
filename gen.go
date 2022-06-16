package generator

import (
	"sync"
	"sync/atomic"
)

type Runner func()

type Generator func() (r Runner, ok bool)

type Pool struct {
	wg      sync.WaitGroup
	g       Generator
	target  uint64
	running uint64
}

func NewPool(g Generator, count int) *Pool {
	if count <= 0 {
		panic("wrong count")
	}
	p := &Pool{
		g:      g,
		target: uint64(count),
	}

	p.wg.Add(1)
	go p.pump()

	return p
}

func (p *Pool) Wait() {
	p.wg.Wait()
}

func (p *Pool) pump() {
	defer p.wg.Done()

	g := p.g

	var task Runner
	var ok bool
	for {
		task, ok = g()
		if !ok {
			return
		}
		newCount := atomic.AddUint64(&p.running, 1)
		if newCount < p.target {
			p.wg.Add(1)
			go p.pump()
		} else {
			goto Enough
		}

		task()
	}

Enough:
	for {
		task()
		task, ok = g()
		if !ok {
			return
		}
	}
}
