package tasks

import (
	"context"
	"fmt"
	"github.com/go-gilbert/gilbert/log"
	"github.com/go-gilbert/gilbert/manifest"
	"github.com/go-gilbert/gilbert/plugins"
	"github.com/go-gilbert/gilbert/runner"
	"github.com/go-gilbert/gilbert/scope"
	"github.com/urfave/cli"
	"os"
	"os/signal"
)

func wrapManifestError(parent error) error {
	return fmt.Errorf("%s\n\nCheck if 'gilbert.yaml' file exists or has correct syntax and check all import statements", parent)
}

// RunTask is a handler for 'run' command
func RunTask(c *cli.Context) (err error) {
	// Read cmd args
	if c.NArg() == 0 {
		return fmt.Errorf("no task specified")
	}

	task := c.Args()[0]

	// Get working dir and read manifest
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("cannot get current working directory, %v", err)
	}

	man, err := manifest.FromDirectory(cwd)
	if err != nil {
		return wrapManifestError(err)
	}

	// Prepare context and import plugins
	ctx, cancelFn := context.WithCancel(context.Background())
	if err := importProjectPlugins(ctx, man, cwd); err != nil {
		return wrapManifestError(err)
	}

	// Run the task
	tr := runner.NewTaskRunner(man, cwd, log.Default)
	go handleShutdown(cancelFn)

	if err := tr.RunTask(task); err != nil {
		return err
	}

	log.Default.Successf("Task '%s' ran successfully\n", task)
	return nil
}

func importProjectPlugins(ctx context.Context, m *manifest.Manifest, cwd string) error {
	s := scope.CreateScope(cwd, m.Vars)
	for _, uri := range m.Plugins {
		expanded, err := s.ExpandVariables(uri)
		if err != nil {
			return fmt.Errorf("failed to load plugins from manifest, %s", err)
		}

		if err := plugins.Import(ctx, expanded); err != nil {
			return err
		}
	}

	return nil
}

func handleShutdown(cancelFn context.CancelFunc) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	for range c {
		log.Default.Log("Shutting down...")
		cancelFn()
	}
}
