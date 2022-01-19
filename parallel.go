package tools

import (
	"fmt"
	"strings"
	"sync"
)

// NewParallelRunner
// 设置并行框架, parallelNum 设置并行数量，如果为0，默认50个并行数
// 设置等待数，到达等待数后停止wait
func NewParallelRunner(parallelNum, waitTotal int) *ParallelRunner {
	if parallelNum <= 0 {
		parallelNum = 50
	}

	p := &ParallelRunner{}
	p.barrier = make(chan struct{}, parallelNum)
	p.errLk = new(sync.Mutex)
	p.errs = nil
	for i := 0; i < parallelNum; i++ {
		p.Release()
	}

	if waitTotal > 0 {
		var wg sync.WaitGroup
		p.wg = &wg
		p.Add(waitTotal)
	}

	return p
}

type Errs []error
type ParallelRunner struct {
	barrier chan struct{}
	wg      *sync.WaitGroup
	errLk   *sync.Mutex
	errs    Errs
}

func (r *ParallelRunner) Acquire() {
	<-r.barrier
}

func (r *ParallelRunner) Release() {
	r.barrier <- struct{}{}
}

func (r *ParallelRunner) Run(lambda func() error) {
	go func() {
		defer r.Done()
		r.Acquire()
		defer r.Release()
		err := lambda()
		if err != nil {
			r.errLk.Lock()
			if r.errs == nil {
				r.errs = make(Errs, 0, 3)
			}
			r.errs = append(r.errs, err)
			r.errLk.Unlock()
		}
	}()
}

func (r *ParallelRunner) Wait() {
	if r.wg != nil {
		r.wg.Wait()
	}
}

func (r *ParallelRunner) Error() error {
	r.errLk.Lock()
	defer r.errLk.Unlock()
	if r.errs == nil || len(r.errs) == 0 {
		return nil
	}

	return r.errs
}

func (r *ParallelRunner) Add(delta int) {
	if r.wg != nil {
		r.wg.Add(delta)
	}
}

func (r *ParallelRunner) Done() {
	if r.wg != nil {
		r.wg.Done()
	}
}

func (e Errs) Error() string {
	if len(e) == 0 {
		return ""
	}

	var causes []string
	for _, err := range e {
		causes = append(causes, err.Error())
	}

	return fmt.Sprintf("TotalError: %d; ", len(e)) + strings.Join(causes, "\n")
}
