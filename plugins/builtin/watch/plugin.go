package watch

import (
	"fmt"
	"github.com/rjeczalik/notify"
	"github.com/x1unix/gilbert/logging"
	"github.com/x1unix/gilbert/plugins"
	"github.com/x1unix/gilbert/runner/job"
	"github.com/x1unix/gilbert/scope"
	"time"
)

type Plugin struct {
	params
	scope  *scope.Scope
	log    logging.Logger
	done   chan bool
	events chan notify.EventInfo
}

func newPlugin(s *scope.Scope, p params, l logging.Logger) (*Plugin, error) {
	return &Plugin{
		params: p,
		scope:  s,
		log:    l,
		done:   make(chan bool),
	}, nil
}

func (p *Plugin) Call(ctx *job.RunContext, r plugins.TaskRunner) error {
	p.events = make(chan notify.EventInfo, 1)
	if err := notify.Watch(p.Path, p.events, notify.All); err != nil {
		return fmt.Errorf("failed to initialize watcher for '%s': %s", p.Path, err)
	}

	childCtx := ctx.ChildContext()
	defer func() {
		notify.Stop(p.events)
		childCtx.Cancel()
		p.log.Debug("watcher removed")
	}()

	go func() {
		interval := p.DebounceTime.ToDuration()
		timer := time.NewTimer(interval)
		hasEvent := false
		for {
			select {
			case event, ok := <-p.events:
				if !ok {
					return
				}
				hasEvent = true
				p.log.Info("event: %v %s", event.Event(), event.Path())
				timer.Reset(interval)
			case <-timer.C:
				if hasEvent {
					hasEvent = false
					childCtx.Cancel()
					childCtx = ctx.ChildContext()
					go p.invokeJob(childCtx, r)
				}
			}
		}
	}()

	p.log.Info("watcher is watching for changes in '%s'", p.Path)
	<-p.done
	return nil
}

func (p *Plugin) invokeJob(ctx job.RunContext, r plugins.TaskRunner) {
	p.log.Debug("job invoke start")
	r.RunJob(p.Job, ctx)
	select {
	case err := <-ctx.Error:
		if err != nil {
			p.log.Error("Error: %s", err)
		}
	}
}

func (p *Plugin) Cancel(ctx *job.RunContext) error {
	p.done <- true
	notify.Stop(p.events)
	p.log.Debug("watcher removed")
	return nil
}
