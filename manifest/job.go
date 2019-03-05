package manifest

import (
	"github.com/x1unix/gilbert/scope"
)

// JobExecType represents job type
type JobExecType uint8

const (
	// ExecEmpty means job has no execution type
	ExecEmpty JobExecType = iota

	// ExecPlugin means that job execute plugin
	ExecPlugin

	// ExecTask means that job runs other task
	ExecTask

	// ExecMixin means that job based on mixin
	ExecMixin
)

// Job is a single job in task
type Job struct {
	// Condition is shell command that should be successful to run specified job
	Condition string `yaml:"if,omitempty"`

	// Description is job description
	Description string `yaml:"description,omitempty"`

	// TaskName refers to task that should be run.
	TaskName *string `yaml:"run,omitempty"`

	// PluginName describes what plugin should handle this job.
	PluginName *string `yaml:"plugin,omitempty"`

	// MixinName is mixin to be used by this job
	MixinName *string `yaml:"mixin,omitempty"`

	// Delay before task start in milliseconds
	Delay uint `yaml:"delay,omitempty"`

	// Vars is a set of variables defined for this job.
	Vars scope.Vars `yaml:"vars,omitempty"`

	// Params is a set of arguments for the job.
	Params map[string]interface{} `yaml:"params,omitempty"`
}

// HasDescription checks if description is available
func (j *Job) HasDescription() bool {
	return j.Description != ""
}

// FormatDescription returns formatted description string
func (j *Job) FormatDescription() string {
	if j.Description != "" {
		return j.Description
	}

	// If description is empty, return used mixin or plugin name if available
	for _, v := range []*string{j.PluginName, j.TaskName, j.MixinName} {
		if v != nil {
			return *v
		}
	}

	return ""
}

// Type returns job execution type
//
// If job has no 'plugin', 'task' or 'plugin' declaration, ExecEmpty will be returned
func (j *Job) Type() JobExecType {
	if j.PluginName != nil {
		return ExecPlugin
	}

	if j.TaskName != nil {
		return ExecTask
	}

	if j.MixinName != nil {
		return ExecMixin
	}

	return ExecEmpty
}
