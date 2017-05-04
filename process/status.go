package process

import (
	"fmt"
	"syscall"
)

type status struct {
	pid     int
	command []string
	status  *syscall.WaitStatus
}

func statusOfPid(pid int, waitStatus *syscall.WaitStatus) status {
	return status{pid: pid, status: waitStatus}
}

func statusOfProcess(process *Process, waitStatus *syscall.WaitStatus) status {
	return status{
		pid: process.Pid, command: process.Command, status: waitStatus}
}

// String returns a formatted string with info about the process.
func (s status) String() string {
	var exitstatus string

	if s.status != nil && s.status.Exited() {
		exitstatus = fmt.Sprintf("status %d", s.status.ExitStatus())
	} else if s.status != nil && s.status.Signaled() {
		exitstatus = fmt.Sprintf("signal %s", s.status.Signal())
	} else {
		exitstatus = "running"
	}

	if len(s.command) == 0 {
		return fmt.Sprintf("process %d, %s", s.pid, exitstatus)
	}
	return fmt.Sprintf("process %d %s, %s", s.pid, s.command, exitstatus)
}

func (s status) isAlive() bool {
	// WIFEXITED(wstatus)
	//   returns true if the child terminated normally, that is, by
	//   calling exit(3) or _exit(2), or by returning from main().
	// WIFSIGNALED(wstatus)
	//   returns true if the child process was terminated by a signal.
	// If neither is true, it's still alive.
	if s.status == nil {
		return true
	}
	return !s.status.Exited() && !s.status.Signaled()
}

// hasFailed is true if the process has ended badly.
func (s status) hasFailed() bool {
	if s.isAlive() {
		return false
	}
	if s.status.Exited() && s.status.ExitStatus() == 0 {
		return false
	}
	// Either a signal or a non-zero exit status.
	return true
}
