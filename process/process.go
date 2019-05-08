package process

import (
	"fmt"
	"os"
	"strings"
	"syscall"
	"time"
)

const (
	// pidDone states that the process ended and we don't want to start
	// anew.  This value doesn't conflict with valid PIDs.
	pidDone = 0
	// pidFailed states that the process ended with a failure state and
	// we *do* want to start anew.  This value doesn't conflict with
	// valid PIDs.
	pidFailed = -1
	// pidValid is the lowest valid PID, which is 1.
	pidValid = 1
)

// Process keeps track of a single subprocess.
type Process struct {
	Command []string
	Pid     int
}

// New creates a new process.
func New(command []string) (Process, error) {
	// Note that we require a non-clean environment. We want the ENV
	// from the caller to end up here.
	process := Process{Command: command, Pid: pidDone}
	err := process.respawn()
	return process, err
}

// Spawn/respawn process.
func (p *Process) respawn() error {
	if p.Pid >= pidValid {
		return &alreadyRunningError{}
	}

	workingDir, err := os.Getwd()
	if err != nil {
		return err
	}
	env := os.Environ()
	files := []uintptr{0, 1, 2} // STDIN, STDOUT, STDERR
	attr := syscall.ProcAttr{Dir: workingDir, Env: env, Files: files}

	pid, err := syscall.ForkExec(searchPathEnv(p.Command[0]), p.Command, &attr)
	if err == nil {
		p.Pid = pid
		fmt.Fprintf(os.Stdout, "Spawned %s\n", statusOfProcess(p, nil))
	}
	return err
}

// Set status of process based on WaitStatus
func (p *Process) setStatus(waitStatus *syscall.WaitStatus) {
	status := statusOfProcess(p, waitStatus)
	if status.isAlive() {
		// We don't really expect status statuss for living children.
		// Did you waitpid with WCONTINUED?
		fmt.Fprintf(os.Stderr, "ERR: Not reaping %s\n", status)
	} else {
		fmt.Fprintf(os.Stdout, "Reaped %s\n", status)
		if status.hasFailed() {
			p.Pid = pidFailed
		} else {
			p.Pid = pidDone
		}
	}
}

// WaitForAllProcesses waits for all known and unknown child processes
// of this process.  Because we have PR_SET_CHILD_SUBREAPER, we may get
// more children than we expect.  Allow us to wait for them when
// shutting down.
func WaitForAllProcesses() {
	var waitStatus syscall.WaitStatus
	for {
		// Infinitely block and wait for all children. We cannot do
		// non-blocking WNOHANG, because then we won't get ECHILD once
		// all processes are stopped/reaped.  (And I have no idea how to
		// get a listing of child processes without resorting to
		// proc(5).)
		pid, err := syscall.Wait4(-1, &waitStatus, 0, nil)
		if err == nil {
			status := statusOfPid(pid, &waitStatus)
			fmt.Fprintf(os.Stdout, "Reaped %s\n", status)
		} else if err == syscall.ECHILD {
			//fmt.Fprintf(os.Stderr, "DBG: No more children\n")
			break
		} else {
			fmt.Fprintf(os.Stderr, "ERR: WaitForAllProcesses: %s\n",
				err.Error())
			time.Sleep(1 * time.Second)
		}
	}
}

type alreadyRunningError struct{}

func (e *alreadyRunningError) Error() string {
	return "already running"
}

func searchPathEnv(command string) string {
	// If there is a slash in the command, then don't search the path.
	if strings.IndexByte(command, '/') != -1 {
		return command
	}

	var paths []string
	if pathEnv, hasPath := os.LookupEnv("PATH"); hasPath {
		paths = strings.Split(pathEnv, ":")
	} else {
		paths = []string{
			"/usr/local/sbin", "/usr/local/bin",
			"/usr/sbin", "/usr/bin", "/sbin", "/bin"}
	}

	return searchPaths(command, paths)
}

func searchPaths(command string, paths []string) string {
	for _, path := range paths {
		if path != "" {
			fullPath := path + "/" + command
			info, err := os.Stat(fullPath)
			if err == nil {
				mode := info.Mode()
				if mode.IsRegular() && mode.Perm()&0111 != 0 {
					return fullPath
				}
			}
		}
	}

	return command
}
