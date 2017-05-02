// fixme
package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/ossobv/gospawn/syslog2stdout"
)

// http://git.suckless.org/sinit/tree/sinit.c
// also todo: remove stale files
// also todo: fix the syslog2stdout functionings
// (pass a function to handleAll to convert the Addr struct)


var signalChan chan os.Signal

func main() {
	var syslogds []syslog2stdout.Listener

	initSignals()

	// For each argv, open up a new UDP/UNIXDGRAM listener to
	// "handle". Once we encounter "--" we break.
	argi := 1
	for ; argi < len(os.Args); argi++ {
		if os.Args[argi] == "--" {
			argi++
			break
		}

		listener, err := syslog2stdout.Listen(os.Args[argi])
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
		} else {
			defer listener.Close()
			syslogds = append(syslogds, listener)
		}
	}

	// Leftover args? Start those.
	argp := argi
	for ; argp < len(os.Args); argp++ {
		if os.Args[argp] == "--" || argp == (len(os.Args) - 1) {
			fmt.Printf("should start %s ..\n", os.Args[argi])
			argi = argp + 1
		}
	}

	// Start all listeners in the background.
	for i := 0; i < len(syslogds); i++ {
		go syslogds[i].HandleAll()
	}

	// Start all other processes...?
	// Write message that we're up and running?
	waitSignals()
}

func initSignals() {
	signalChan = make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGHUP)
	signal.Notify(signalChan, syscall.SIGINT)
	signal.Notify(signalChan, syscall.SIGQUIT)		// status report?
	//signal.Notify(signalChan, syscall.SIGUSR1)	// reload?
	//signal.Notify(signalChan, syscall.SIGUSR2)	// reload?
	signal.Notify(signalChan, syscall.SIGTERM)
	//signal.Notify(signalChan, syscall.SIGCHLD)	// useful?
	//wg := &sync.WaitGroup{}
}

// waitSignals waits forever, until a signal of INT or TERM arrives.
// This is better than using a "select{}" as blockForever, because if we
// never return to main, we won't call our deferred Close()s.
func waitSignals() {
	for sig := range signalChan {
		switch sig.String() {
		case "interrupt": fallthrough
		case "term": fallthrough
		case "whatever": return
		default: fmt.Fprintf(os.Stderr, "signal: %s (ignoring)\n", sig)
		}
	}
}
