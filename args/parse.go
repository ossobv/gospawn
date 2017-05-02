package args

// Args holds the user-supplied configuration.
type Args struct {
	// SyslogPorts holds which ports/paths should be listened on.
	SyslogPorts []string
	// Commands holds a list of commands to run/watch.
	Commands [][]string
}

// Parse parses the (command line) args into an Args structure.
func Parse(args []string) (Args) {
	listsOfLists := split(args)

	var ports []string
	var commands [][]string

	if len(args) > 0 {
		ports = listsOfLists[0]
		if len(args) > 1 {
			commands = listsOfLists[1:]
		}
	}

	return Args{SyslogPorts: ports, Commands: commands}
}

// split parses the (command line) args into a list of lists.
// 
// For example:
//   split(["a", "--", "b", "c"]) == [["a"], ["b", "c"]]
func split(args []string) ([][]string) {
	var output [][]string

	start, p := 0, 0
	for ; p < len(args); p++ {
		if args[p] == "--" {
			if p - start > 0 {
				output = append(output, args[start:p])
			}
			start = p + 1
		}
	}
	if p - start > 0 {
		output = append(output, args[start:p])
	}
	return output
}
