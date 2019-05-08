// fixme
package main

import (
	"fmt"
	"os"
	"syscall"

	"github.com/ossobv/gospawn/args"
	"github.com/ossobv/gospawn/prctl"
	"github.com/ossobv/gospawn/process"
	"github.com/ossobv/gospawn/signal"
	"github.com/ossobv/gospawn/syslog2stdout"
)

const (
	// quitImmediately toggles whether CTRL+\\ (SIGQUIT) aborts the
	// application instead of passing the signal on as usual.  Useful
	// during development.  Otherwise only SIGINT and SIGTERM schedule
	// application termination.
	quitImmediately = true
	// sleepBeforeRespawn defines after how many seconds the "respawn all
	// processes" alarm should be fired.
	sleepBeforeRespawn = 10
)

type mainState struct {
	// A list of syslogds.
	syslogds []syslog2stdout.Syslogd
	// A list of processes.
	processlist process.List
	// Fetch signals from here.
	sigHandler signal.Handler
	// No respawning when we're stopping; defaults to false.
	stopping bool
}

func goSpawn() mainState {
	return mainState{stopping: false}
}

// Make us the parent of all children, even daemonized ones.  This fixes
// so we can keep os.Wait()ing for double-forked daemons.
//
// NOTE: This is NOT needed when gospawn is run as Docker PID 1. Docker
// already sets that for us.  But when we want to test stuff without
// Docker this is a blessing.
func (m *mainState) initSubreaper() {
	err := prctl.SetChildSubreaper()
	if err != nil {
		fmt.Fprintf(os.Stderr, "WARN: %s\n", err.Error())
	}
}

func (m *mainState) initSignals() {
	m.sigHandler = signal.New()
}

func (m *mainState) startSyslogds(portsOrPaths []string) {
	// Open up a new UDP/UNIXDGRAM listener for each syslogport.
	for _, port := range portsOrPaths {
		listener, err := syslog2stdout.New(port)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
		} else {
			m.syslogds = append(m.syslogds, listener)
		}
	}

	// Start all syslogds in the background.
	for i := 0; i < len(m.syslogds); i++ {
		go m.syslogds[i].HandleAll()
	}
}

func (m *mainState) startProcesses(commands [][]string) {
	// Start all commands in the background.
	m.processlist = process.NewList()
	for _, command := range commands {
		err := m.processlist.StartProcess(command)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERR: starting %s: %s\n",
				command, err.Error())
			continue
		}
	}
}

func (m *mainState) stopSyslogds() {
	// Close all syslogds (with cleanup)
	for i := 0; i < len(m.syslogds); i++ {
		m.syslogds[i].Close()
	}
}

func (m *mainState) stopProcesses() {
	// Kill -9 all processes that hadn't been killed yet.
	m.processlist.SendSignal(syscall.SIGKILL)
}

func (m *mainState) shutdown() {
	m.stopProcesses()
	m.stopSyslogds()
}

func (m *mainState) hasWork() bool {
	return (len(m.syslogds) != 0 ||
		!m.processlist.IsEmpty())
}

func (m *mainState) handleAlarm() {
	// We send ourself an alarm when there has been a change in
	// the processlist.  Respawn processes that have died.
	if !m.stopping {
		m.processlist.RespawnFailed()
	}
}

func (m *mainState) handleSigChild() {
	didSomething := false
	for m.processlist.HandleSigChild() {
		didSomething = true
		// Maybe we only get one SIGCHLD for multiple children. Reap as
		// much as possible.
		for m.processlist.HandleSigChild() == true {
		}
	}
	if !m.processlist.IsEmpty() && m.processlist.IsDone() {
		// If we're running processes, but they're all done
		// (completed with success code), then we can stop.
		m.stopping = true
	} else if didSomething {
		// If we did something to the process list, we may have
		// lost a child. Spawn an alarm to restart any dead
		// children soon.
		signal.Alarm(sleepBeforeRespawn)
	}
}

func (m *mainState) handleQuit() {
	if quitImmediately {
		fmt.Fprintf(os.Stderr,
			"ERR: Got SIGQUIT, passing kill -9 to all\n")
		// Quick exit, no cleanup!
		m.processlist.SendSignal(syscall.SIGKILL)
		// Try a bit of cleanup.
		m.shutdown()
		os.Exit(128 + 3 /* SIGQUIT */)
	}
}

func (m *mainState) handleSigDefault(sig os.Signal) {
	sigDesc := sig.String()
	if sigDesc == "interrupt" || sigDesc == "terminated" {
		m.stopping = true
	}

	if syscallSig, ok := sig.(syscall.Signal); ok {
		fmt.Fprintf(os.Stderr, "DBG: Forwarding signal %s\n",
			syscallSig)
		m.processlist.SendSignal(syscallSig)
	}

	if sigDesc == "stopped" {
		// Background self because of SIGTSTP.
		syscall.Kill(syscall.Getpid(), syscall.SIGSTOP)
	}
}

func (m *mainState) doWork() {
	// Read all signals until it's time to end.
	for sig := range m.sigHandler.Chan {
		//fmt.Fprintf(os.Stderr, "DBG: Signal %s\n", sig.String())
		m.doHandler(sig, true)

		// If we're "stopping" and there are no more running processes,
		// we're done.
		if m.stopping && !m.processlist.IsRunning() {
			break
		}
	}
}

func (m *mainState) doHandler(sig os.Signal, handleChildren bool) {
	switch sig.String() {
	case "alarm clock":
		m.handleAlarm()

	case "child exited":
		if handleChildren == true {
			m.handleSigChild()
		}

	case "quit":
		// handleQuit() may Exit, or not, in which case we handle it
		// like every other signal.
		m.handleQuit()
		fallthrough

	default:
		m.handleSigDefault(sig)
	}
}

func main() {
	args := args.Parse(os.Args[1:])

	gospawn := goSpawn()
	gospawn.initSubreaper()
	gospawn.initSignals()
	gospawn.startSyslogds(args.SyslogPorts)
	gospawn.startProcesses(args.Commands)
	if gospawn.hasWork() {
		gospawn.doWork()
	}
	gospawn.shutdown()
}
