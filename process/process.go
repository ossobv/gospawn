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

	workingDir, err := os.Getwd()
	if err != nil {
		return err
	}
	env := os.Environ()
	files := []uintptr{0, 1, 2}
	attr := syscall.ProcAttr{Dir: workingDir, Env: env, Files: files}

	pid, err := syscall.ForkExec(p.Command[0], p.Command, &attr)
	if err == nil {
		fmt.Fprintf(os.Stdout, "Spawned process %d: %s\n", pid, p.Command)
		p.Pid = pid
	}
	return err
}

// Set status of process based on WaitStatus
func (p *Process) setStatus(waitStatus *syscall.WaitStatus) {
	if waitStatus.Exited() {
		fmt.Fprintf(os.Stdout, "Reaped process %d: %s, status %d\n",
				p.Pid, p.Command, waitStatus.ExitStatus())
		if waitStatus.ExitStatus() == 0 {
			p.Pid = PID_DONE
		} else {
			p.Pid = PID_FAILED
		}
	} else if waitStatus.Signaled() {
		fmt.Fprintf(os.Stdout, "Reaped process %d: %s, signal %s\n",
				p.Pid, p.Command, waitStatus.Signal())
		p.Pid = PID_FAILED
	} else {
		fmt.Fprintf(os.Stderr, "DBG: Not reaping PID %d\n", p.Pid)
	}
}

type alreadyRunningError struct {}

func (e *alreadyRunningError) Error() string {
	return "already running"
}
