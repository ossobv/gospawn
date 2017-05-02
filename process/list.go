package process

import (
	"fmt"
	"os"
	"syscall"
)

// List holds a list of processes to work on with predefined methods.
type List struct {
	processes []Process
}

// NewList returns a new empty list of processes to work on.
func NewList() List {
	return List{}
}

// Add adds a process to the list of processes.
func (l *List) Add(proc Process) {
	l.processes = append(l.processes, proc)
}

// SendSignal sends a signal to all running processes in the list.
func (l *List) SendSignal(signal syscall.Signal) {
	for i := 0; i < len(l.processes); i++ {
		if l.processes[i].Pid >= PID_VALID {
			syscall.Kill(l.processes[i].Pid, signal)
		}
	}
}

// IsDone returns whether the processes are all done without failure.
func (l *List) IsDone() bool {
	for i := 0; i < len(l.processes); i++ {
		if l.processes[i].Pid != PID_DONE {
			return false
		}
	}
	return true
}

// IsEmpty returns whether the process list is empty.
func (l *List) IsEmpty() bool {
	return len(l.processes) == 0
}

// IsRunning returns whether any process is currently running.
func (l *List) IsRunning() bool {
	for i := 0; i < len(l.processes); i++ {
		if l.processes[i].Pid >= PID_VALID {
			//fmt.Fprintf(os.Stderr, "DBG: Methinks PID %d is still up...\n",
			//		l.processes[i].Pid)
			return true
		}
	}
	return false
}

// RespawnFailed respawns all processes that have a failure exit code.
func (l *List) RespawnFailed() uint {
	count := uint(0)
	for i := 0; i < len(l.processes); i++ {
		if l.processes[i].Pid == PID_FAILED {
			err := l.processes[i].respawn()
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
    var w syscall.WaitStatus
	pid, err := syscall.Wait4(-1, &w, syscall.WNOHANG, nil)

	// Nothing to do?
	if pid == 0 {
		return false
	}

	// In the rare case that we missed a signal, we can use the "there
	// are no processes to wait on" ECHILD to mark all children down.
	if err == syscall.ECHILD {
		for i := 0; i < len(l.processes); i++ {
			if l.processes[i].Pid >= PID_VALID {
				fmt.Fprintf(os.Stderr,
						"ERR: acting on ECHILD, marking PID %d as down\n",
						l.processes[i].Pid)
				l.processes[i].Pid = PID_FAILED
			}
		}
		return false
	}

	// Error, pretend something happened so we can double-check in case
	// there is an EAGAIN/EINTR or something.
	if pid < 0 {
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERR: wait4: %s\n", err.Error())
		}
		return true
	}

	// Looping over _, proc := range processes is no good.
	// It'll return a copy of Process, and we'd edit a temp
	// copy only.
	for i := 0; i < len(l.processes); i++ {
		if l.processes[i].Pid == pid {
			if w.Exited() {
				fmt.Fprintf(os.Stdout, "Reaped process %d: %s, status %d\n",
						pid, l.processes[i].Command, w.ExitStatus())
				if w.ExitStatus() == 0 {
					l.processes[i].Pid = PID_DONE
				} else {
					l.processes[i].Pid = PID_FAILED
				}
			} else if w.Signaled() {
				fmt.Fprintf(os.Stdout, "Reaped process %d: %s, signal %s\n",
						pid, l.processes[i].Command, w.Signal())
				l.processes[i].Pid = PID_FAILED
			} else {
				fmt.Fprintf(os.Stderr, "DBG: Not reaping PID %d\n", pid)
			}
			return true
		}
	}

	// Apparently this was not one of our children.
	return false
}
