package signal

import (
	"os"
	"os/signal"
	"syscall"
)

const (
	// Signals in 32..64 are realtime signals.  Portability to non-Linux
	// platforms is not of our concern at this time.
	MAXSIG = 31
)

// sigHandler is used to handle the handle INT, TERM and CHLD signals.
type sigHandler struct {
	Chan chan os.Signal
}

// New initializes the signal handlers.  We now get events for all
// signals.  Read the sigHandler.Chan and handle them appropriately.
func New() sigHandler {
	signalChan := make(chan os.Signal, 1)
	for i := 1; i <= MAXSIG; i++ {
		signal.Notify(signalChan, syscall.Signal(i))
	}
	return sigHandler{Chan: signalChan}
}
