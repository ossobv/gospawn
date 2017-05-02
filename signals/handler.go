package signals

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

// SigHandler is used to handle the handle INT, TERM and CHLD signals.
type SigHandler struct {
	signalChan chan os.Signal
}

// New initializes the signal handlers.  Calling HandleAll will make it
// block and listen to all signals.
func New() SigHandler {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGHUP)
	signal.Notify(signalChan, syscall.SIGINT)
	signal.Notify(signalChan, syscall.SIGQUIT)		// status report?
	//signal.Notify(signalChan, syscall.SIGUSR1)	// reload?
	//signal.Notify(signalChan, syscall.SIGUSR2)	// reload?
	signal.Notify(signalChan, syscall.SIGTERM)
	//signal.Notify(signalChan, syscall.SIGCHLD)	// useful?
	return SigHandler{signalChan: signalChan}
}

// HandleAll waits forever, until a signal of INT or TERM arrives.  This
// is better than using a "select{}" as blockForever, because if we
// never return to main, we won't call any deferred Close()s. (Same with
// an early os.Exit().)
func (h *SigHandler) HandleAll() {
	for sig := range h.signalChan {
		switch sig.String() {
			// Handle "interrupt" (SIGINT) and "term" (SIGTERM) by
			// returning.
			case "interrupt": fallthrough
			case "term": fallthrough
			case "...": return
			// Show a temporary message for the others.
			default: fmt.Fprintf(os.Stderr, "signal: %s (ignoring)\n", sig)
		}
	}
}
