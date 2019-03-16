package tasks

import (
	"fmt"
	"github.com/urfave/cli"
	"github.com/x1unix/gilbert/logging"
	"github.com/x1unix/gilbert/manifest"
	"github.com/x1unix/gilbert/runner"
	"os"
	"os/signal"
)

var (
	r *runner.TaskRunner
)

func getManifest(dir string) (*manifest.Manifest, error) {
	dir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("cannot get current working directory, %v", err)
	}

	return manifest.FromDirectory(dir)
}

// RunTask is a handler for 'run' command
func RunTask(c *cli.Context) (err error) {
	if c.NArg() == 0 {
		return fmt.Errorf("no task specified")
	}

	task := c.Args()[0]

	r, err = getRunner()
	if err != nil {
		return err
	}

	return runTask(task)
}

func getRunner() (*runner.TaskRunner, error) {
	dir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("cannot get current working directory, %v", err)
	}

	m, err := getManifest(dir)
	if err != nil {
		return nil, err
	}

	return runner.NewTaskRunner(m, dir, log.Log), nil
}

func handleShutdown() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	for range c {
		log.Log.Log("Shutting down...")
		r.Stop()
	}
}

func runTask(taskName string) error {
	go handleShutdown()
	if err := r.RunTask(taskName); err != nil {
		return err
	}

	log.Log.Success("Task '%s' ran successfully\n", taskName)
	return nil
}
