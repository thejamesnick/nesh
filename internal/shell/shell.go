// Package shell is the OS adapter seam.
//
// Contract (MODULES.md): the ONLY package allowed to import os, os/exec,
// syscall, net. The runtime speaks to the interfaces defined here; tests
// inject fakes.
package shell

import (
	"fmt"
	"io"
	"os/exec"
)

// CommandRunner spawns processes and captures exit codes.
type CommandRunner interface {
	// Run executes name with args. Output streams into the provided
	// writers. It returns the process exit code (127 when the command
	// cannot be found or started).
	Run(name string, args []string, stdout, stderr io.Writer) int
}

// RealRunner runs commands from PATH via os/exec.
type RealRunner struct{}

// Run implements CommandRunner.
func (RealRunner) Run(name string, args []string, stdout, stderr io.Writer) int {
	cmd := exec.Command(name, args...) // #nosec G204 — by design: nesh runs user commands
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	if err := cmd.Run(); err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			return ee.ExitCode()
		}
		fmt.Fprintf(stderr, "nesh: %v\n", err)
		return 127
	}
	return 0
}
