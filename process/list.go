package process

import (
	"fmt"
	"os"
	"syscall"
)

// List holds a list of processes to work on with predefined methods.
type List []Process

// NewList returns a new empty list of processes to work on.
func NewList() List {
	return List{}
}

// StartProcess starts a process and adds it to the list of processes,
// unless it failed.
func (l *List) StartProcess(command []string) error {
	proc, err := New(command)
	if err != nil {
		return err
	}
	*l = append(*l, proc)
	return nil
}

// SendSignal sends a signal to all running processes in the list.
func (l *List) SendSignal(signal syscall.Signal) {
	for i := 0; i < len(*l); i++ {
		if (*l)[i].Pid >= PID_VALID {
			syscall.Kill((*l)[i].Pid, signal)
		}
	}
}

// IsDone returns whether the processes are all done without failure.
func (l *List) IsDone() bool {
	for i := 0; i < len(*l); i++ {
		if (*l)[i].Pid != PID_DONE {
			return false
		}
	}
	return true
}

// IsEmpty returns whether the process list is empty.
func (l *List) IsEmpty() bool {
	return len(*l) == 0
}

// IsRunning returns whether any process is currently running.
func (l *List) IsRunning() bool {
	for i := 0; i < len(*l); i++ {
		if (*l)[i].Pid >= PID_VALID {
			//fmt.Fprintf(os.Stderr, "DBG: Methinks PID %d is still up...\n",
			//		(*l)[i].Pid)
			return true
		}
	}
	return false
}

// RespawnFailed respawns all processes that have a failure exit code.
func (l *List) RespawnFailed() uint {
	count := uint(0)
	for i := 0; i < len(*l); i++ {
		if (*l)[i].Pid == PID_FAILED {
			err := (*l)[i].respawn()
			if err == nil {
				count++
			}
		}
	}
	return count
}

// HandleSigChild calls waitpid and marks processes from the process
// list as done.  Return true if there was something to handle.
func (l *List) HandleSigChild() bool {
	var waitStatus syscall.WaitStatus
	pid, err := syscall.Wait4(-1, &waitStatus, syscall.WNOHANG, nil)

	switch {
	case pid == 0:
		// Nothing to do?
		return false

	case err == syscall.ECHILD:
		// In the rare case that we missed a signal, we can use the
		// "there are no processes to wait on" ECHILD to mark all
		// children down.
		for i := 0; i < len(*l); i++ {
			if (*l)[i].Pid >= PID_VALID {
				fmt.Fprintf(os.Stderr,
					"ERR: acting on ECHILD, marking PID %d as down\n",
					(*l)[i].Pid)
				(*l)[i].Pid = PID_FAILED
			}
		}
		return false

	case pid < 0:
		// Error, pretend something happened so we can double-check in
		// case there is an EAGAIN/EINTR or something.
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERR: wait4: %s\n", err.Error())
		}
		return true

	default:
		// Looping over _, proc := range processes is no good.
		// It'll return a copy of Process, and we'd edit a temp
		// copy only.
		for i := 0; i < len(*l); i++ {
			if (*l)[i].Pid == pid {
				(*l)[i].setStatus(&waitStatus)
				return true
			}
		}
		fmt.Fprintf(os.Stdout, "Reaped %s\n", statusOfPid(pid, &waitStatus))
	}

	// Apparently this was not one of our children.
	return false
}
