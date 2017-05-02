package process

import (
	"fmt"
	"os"
    "syscall"
)

const (
	// PID_DONE states that the process ended and we don't want to start
	// anew.  This value doesn't conflict with valid PIDs.
	PID_DONE = 0
	// PID_FAILED states that the process ended with a failure state and
	// we *do* want to start anew.  This value doesn't conflict with
	// valid PIDs.
	PID_FAILED = -1
	// PID_VALID is the lowest valid PID, which is 1.
	PID_VALID = 1
)

// Process keeps track of a single subprocess.
type Process struct {
	Command []string
	Pid int
}

// New creates a new process.
func New(command []string) (Process, error) {
	// Note that we require a non-clean environment. We want the ENV
	// from the caller to end up here.
	process := Process{Command: command, Pid: PID_DONE}
	err := process.respawn()
	return process, err
}

// Spawn/respawn process.
func (p *Process) respawn() error {
	if p.Pid >= PID_VALID {
		return &alreadyRunningError{}
	}

	pid, err := syscall.ForkExec(p.Command[0], p.Command, nil)
	if err == nil {
		fmt.Fprintf(os.Stdout, "Spawned process %d: %s\n", pid, p.Command)
		p.Pid = pid
	}
	return err
}

type alreadyRunningError struct {}

func (e *alreadyRunningError) Error() string {
	return "already running"
}
