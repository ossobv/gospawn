// fixme
package main

import (
	"fmt"
	"os"
	"syscall"

	"github.com/ossobv/gospawn/args"
	"github.com/ossobv/gospawn/process"
	"github.com/ossobv/gospawn/signal"
	"github.com/ossobv/gospawn/syslog2stdout"
)

const (
	// QUIT_IMMEDIATELY toggles whether CTRL+\\ (SIGQUIT) aborts the
	// application instead of passing the signal on as usual.  Useful
	// during development.  Otherwise only SIGINT and SIGTERM schedule
	// application termination.
	QUIT_IMMEDIATELY = true
	// SLEEP_BEFORE_RESPAWN defines after how many seconds the "respawn all
	// processes" alarm should be fired.
	SLEEP_BEFORE_RESPAWN = 10
)

func main() {
	var syslogds []syslog2stdout.Syslogd

	args := args.Parse(os.Args[1:])

	sigHandler := signal.New()

	// Open up a new UDP/UNIXDGRAM listener for each syslogport.
	for _, port := range args.SyslogPorts {
		listener, err := syslog2stdout.New(port)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
		} else {
			defer listener.Close()
			syslogds = append(syslogds, listener)
		}
	}

	// Start all syslogds in the background.
	for _, syslogd := range syslogds {
		go syslogd.HandleAll()
	}

	// Start all commands in the background.
	processlist := process.NewList()
	for _, command := range args.Commands {
		process, err := process.New(command)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERR: starting %s: %s\n",
					command, err.Error())
			continue
		}
		processlist.Add(process)
	}

	// If there are no syslogds, and no processes either, then we're
	// done here.
	if len(syslogds) == 0 && processlist.IsEmpty() {
		os.Exit(0)
	}

	// If we're stopping, then don't respawn anything.
	stopping := false

	// Read all signals until it's time to end.
	for sig := range sigHandler.Chan {
		sigDesc := sig.String()
		//fmt.Fprintf(os.Stderr, "DBG: signal: %s\n", sigDesc)

		switch sigDesc {
		case "alarm clock":
			// We send ourself an alarm when there has been a change in
			// the processlist.  Respawn processes that have died.
			if !stopping {
				processlist.RespawnFailed()
			}

		case "child exited":
			didSomething := false
			for ; processlist.HandleSigChild(); {
				didSomething = true
			}
			if !processlist.IsEmpty() && processlist.IsDone() {
				// If we're running processes, but they're all done
				// (completed with success code), then we can stop.
				stopping = true
			} else if didSomething {
				// If we did something to the process list, we may have
				// lost a child. Spawn an alarm to restart any dead
				// children soon.
				signal.Alarm(SLEEP_BEFORE_RESPAWN)
			}

		case "quit":
			if QUIT_IMMEDIATELY {
				fmt.Fprintf(os.Stderr,
						"ERR: Got SIGQUIT, passing kill -9 to all\n")
			    // Quick exit, no cleanup!
			    processlist.SendSignal(syscall.SIGKILL)
			    // No deferred Close() handlers will get called.
			    os.Exit(128 + 3 /* SIGQUIT */)
			}
			fallthrough

		default:
			if sigDesc == "interrupt" || sigDesc == "terminated" {
				stopping = true
			}

			if syscallSig, ok := sig.(syscall.Signal); ok {
				//fmt.Fprintf(os.Stderr, "DBG: forwarding signal %s\n",
				//		syscallSig)
				processlist.SendSignal(syscallSig)
			}

			if sigDesc == "stopped" {
				// Background self because of SIGTSTP.
				syscall.Kill(syscall.Getpid(), syscall.SIGSTOP)
			}
		}

		// If we're stopping and there is no running process, then we're done.
		if stopping && !processlist.IsRunning() {
			break
		}
	}
}
