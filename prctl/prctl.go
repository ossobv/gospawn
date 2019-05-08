package prctl

import (
	"fmt"
	"syscall"
)

// Very linux-3.4+-amd64 specific. This will likely fail everywhere else.
const sysPrctl = 157           // SYS_PRCTL
const prSetChildSubreaper = 36 // PR_SET_CHILD_SUBREAPER

// SetChildSubreaper explicitly sets PR_SET_CHILD_SUBREAPER so children
// of this process (including daemonized processes) get to see and reap
// all children.
func SetChildSubreaper() error {
	_, _, ret := syscall.RawSyscall(sysPrctl, prSetChildSubreaper, 1, 0)
	if ret != 0 {
		return fmt.Errorf("PR_SET_CHILD_SUBREAPER failed with errno %d", ret)
	}
	return nil
}
