package job

import (
	"context"
	"sync"
	"time"

	"github.com/x1unix/gilbert/logging"
	"github.com/x1unix/gilbert/scope"
)

type RunContext struct {
	RootVars scope.Vars
	Logger   logging.Logger
	Context  context.Context
	Error    chan error
	child    bool
	wg       *sync.WaitGroup
	once     sync.Once
	finished bool
	cancelFn context.CancelFunc
}

func (r *RunContext) SetWaitGroup(wg *sync.WaitGroup) {
	r.wg = wg
}

func (r *RunContext) IsChild() bool {
	return r.child
}

func (r *RunContext) ForkContext() RunContext {
	return RunContext{
		RootVars: r.RootVars,
		Logger:   r.Logger.SubLogger(),
		Context:  r.Context,
		Error:    r.Error,
		cancelFn: r.cancelFn,
		child:    true,
		wg:       r.wg,
	}
}

func (r *RunContext) ChildContext() RunContext {
	ctx, cancelFn := context.WithCancel(r.Context)

	return RunContext{
		RootVars: r.RootVars,
		Logger:   r.Logger.SubLogger(),
		Context:  ctx,
		Error:    r.Error,
		cancelFn: cancelFn,
		child:    true,
		wg:       r.wg,
	}
}

func (r *RunContext) WithTimeout(t time.Duration) RunContext {
	ctx, fn := context.WithTimeout(r.Context, t)
	return RunContext{
		RootVars: r.RootVars,
		Logger:   r.Logger,
		Context:  ctx,
		cancelFn: fn,
	}
}

func (r *RunContext) Cancel() {
	if r.cancelFn == nil {
		r.Logger.Error("Bug: context cancel function is nil")
		return
	}

	r.cancelFn()
}

func (r *RunContext) Success() {
	r.Result(nil)
}

func (r *RunContext) IsAlive() bool {
	return !r.finished
}

func (r *RunContext) Result(err error) {
	r.once.Do(func() {
		defer func() {
			if rec := recover(); rec != nil {
				r.Logger.Error("Bug: failed to return job result, %v", rec)
				r.wg.Done()
			}
		}()

		r.Error <- err
		if r.wg == nil {
			r.Logger.Warn("Warning: waitgroup was no defined for job context")
			return
		}

		r.finished = true
		r.Logger.Debug("result received")
		r.wg.Done()
	})
}

func NewRunContext(rootVars scope.Vars, log logging.Logger, ctx context.Context) RunContext {
	return RunContext{RootVars: rootVars, Logger: log, Context: ctx, Error: make(chan error, 1)}
}
