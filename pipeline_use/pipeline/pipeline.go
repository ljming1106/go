package pipeline

import (
	"log"
	"sync"
)

func HasClosed(c <-chan struct{}) bool {
	select {
	case <-c:
		return true
	default:
		return false
	}
}

type SyncFlag interface {
	Wait()
	Chan() <-chan struct{}
	Done() bool
}

func NewSyncFlag() (done func(), flag SyncFlag) {
	f := &syncFlag{
		c: make(chan struct{}),
	}
	return f.done, f
}

type syncFlag struct {
	once sync.Once
	c    chan struct{}
}

func (f *syncFlag) done() {
	f.once.Do(func() {
		close(f.c)
	})
}

func (f *syncFlag) Wait() {
	<-f.c
}

func (f *syncFlag) Chan() <-chan struct{} {
	return f.c
}

func (f *syncFlag) Done() bool {
	return HasClosed(f.c)
}

type pipelineThread struct {
	sigs         []chan struct{} //
	chanExit     chan struct{}   // 某工序逻辑结束
	interrupt    SyncFlag        // 中断操作
	setInterrupt func()          // 关闭syncFlag的通道，只执行一次（sync.Once.Do）
	err          error
}

func newPipelineThread(l int) *pipelineThread {
	p := &pipelineThread{
		sigs:     make([]chan struct{}, l),
		chanExit: make(chan struct{}),
	}
	p.setInterrupt, p.interrupt = NewSyncFlag()

	for i := range p.sigs {
		p.sigs[i] = make(chan struct{})
	}
	return p
}

type Pipeline struct {
	mtx         sync.Mutex
	workerChans []chan struct{} //某工序并发数
	prevThd     *pipelineThread
}

func NewPipeline(workers ...int) *Pipeline {
	if len(workers) < 1 {
		panic("NewPipeline need aleast one argument")
	}

	workersChan := make([]chan struct{}, len(workers))
	for i := range workersChan {
		workersChan[i] = make(chan struct{}, workers[i])
	}

	return &Pipeline{
		workerChans: workersChan,
	}
}

func (p *Pipeline) Async(works ...func() error) bool {
	if len(works) != len(p.workerChans) {
		panic("Async: arguments number not matched to NewPipeline(...)")
	}

	p.mtx.Lock()
	thisThd := newPipelineThread(len(p.workerChans))
	p.prevThd = thisThd
	p.mtx.Unlock()

	lock := func(idx int) bool {
		log.Printf("idx[%v] ==> [ len:%v , cap:%v ]", idx, len(p.workerChans[idx]), cap(p.workerChans[idx]))
		select {
		// 如果不堵塞，就说明该工序还空闲
		case p.workerChans[idx] <- struct{}{}: //get lock
		}
		return true
	}
	if !lock(0) {
		close(thisThd.chanExit)
		return false
	}
	go func() {
		var err error
		for i, work := range works {
			// ？？？这个跟chanExit有什么区别
			// sigs ： 开始信号？
			// workerChans ： 结束信号？
			close(thisThd.sigs[i]) //signal next thread
			if work != nil {
				err = work()
			}
			if err != nil || (i+1 < len(works) && !lock(i+1)) {
				thisThd.setInterrupt()
				break
			}
			<-p.workerChans[i] //release lock
		}

		thisThd.err = err
		close(thisThd.chanExit)
	}()
	return true
}

func (p *Pipeline) Wait() error {
	p.mtx.Lock()
	lastThd := p.prevThd
	p.mtx.Unlock()
	<-lastThd.chanExit
	return lastThd.err
}
