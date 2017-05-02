// fixme
package main

import (
	"fmt"
	"os"

	"github.com/ossobv/gospawn/args"
	"github.com/ossobv/gospawn/signals"
	"github.com/ossobv/gospawn/syslog2stdout"
)

// http://git.suckless.org/sinit/tree/sinit.c
// also todo: remove stale files
// also todo: fix the syslog2stdout functionings
// (pass a function to handleAll to convert the Addr struct)


func main() {
	var syslogds []syslog2stdout.Syslogd

	args := args.Parse(os.Args[1:])

	sigHandler := signals.New()

	// Open up a new UDP/UNIXDGRAM listener for each syslogport.
	for _, port := range args.SyslogPorts {
		fmt.Printf("LOG port/path %s\n", port)
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
	for _, command := range args.Commands {
		fmt.Printf("CMD %r\n", command)
	}

	// Start all other processes...?
	// Write message that we're up and running?
	sigHandler.HandleAll()
}
