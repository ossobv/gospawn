package prctl

import (
	"os"
	"syscall"
	"testing"
)

func runSleepInBackground() int {
	env := os.Environ()
	files := []uintptr{} // no fds
	attr := syscall.ProcAttr{Dir: "/", Env: env, Files: files}
	pid, err := syscall.ForkExec("/bin/sh",
		[]string{"/bin/sh", "-c", "sleep 2 >/dev/null 2>&1 </dev/null &"},
		&attr)
	if err != nil {
		return 0
	}
	return pid
}

func TestNoSubreaper(t *testing.T) {
	childPid := runSleepInBackground()
	if childPid == 0 {
		t.Errorf("Got 0 for child pid?")
		return
	}

	var waitStatus syscall.WaitStatus

	if _, err := syscall.Wait4(childPid, &waitStatus, 0, nil); err != nil {
		t.Errorf("Wait4: %s", err.Error())
		return
	}

	if _, err := syscall.Wait4(-1, &waitStatus, 0, nil); err == nil {
		t.Errorf("Wait4: Expected error on second Wait4")
		return
	}
}

func TestSubreaper(t *testing.T) {
	SetChildSubreaper()
	defer unsetChildSubreaper()

	childPid := runSleepInBackground()
	if childPid == 0 {
		t.Errorf("Got 0 for child pid?")
		return
	}

	var waitStatus syscall.WaitStatus

	if _, err := syscall.Wait4(childPid, &waitStatus, 0, nil); err != nil {
		t.Errorf("Wait4: %s", err.Error())
		return
	}

	if _, err := syscall.Wait4(-1, &waitStatus, 0, nil); err != nil {
		t.Errorf("Wait4: Unexpected error on second Wait4")
		return
	}
}
