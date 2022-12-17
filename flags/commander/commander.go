// Package subcommands implements a simple way for a single command to have many
// subcommands, each of which takes arguments and so forth.
package commander

import (
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"text/tabwriter"
)

// An ExitStatus represents a Posix exit status that a subcommand
// expects to be returned to the shell.
type ExitStatus int

const (
	ExitSuccess ExitStatus = iota
	ExitFailure
	ExitUsageError
)

// A Command represents a single command.
type Command interface {
	// Name returns the name of the command.
	Name() string

	// Synopsis returns a short string (less than one line) describing the command.
	Synopsis() string

	// Usage returns a long string explaining the command and giving usage
	// information.
	Usage() string

	// SetFlags adds the flags for this command to the specified set.
	SetFlags(*flag.FlagSet)

	// Execute executes the command and returns an ExitStatus.
	Execute(f *flag.FlagSet) ExitStatus
}

// A CommandGroup represents a set of commands about a common topic.
type CommandGroup struct {
	name     string
	commands []Command
}

// A Commander represents a set of commands.
type Commander struct {
	commands  []*CommandGroup
	topFlags  *flag.FlagSet // top-level flags
	important []string      // important top-level flags
	name      string        // normally path.Base(os.Args[0])

	Explain        func(io.Writer)                // A function to print a top level usage explanation. Can be overridden.
	ExplainGroup   func(io.Writer, *CommandGroup) // A function to print a command group's usage explanation. Can be overridden.
	ExplainCommand func(io.Writer, Command)       // A function to print a command usage explanation. Can be overridden.

	Output io.Writer // Output specifies where the commander should write its output (default: os.Stdout).
	Error  io.Writer // Error specifies where the commander should write its error (default: os.Stderr).
}

// Name returns the group name
func (g *CommandGroup) Name() string {
	return g.name
}

// New returns a new commander with the specified top-level
// flags and command name. The Usage function for the topLevelFlags
// will be set as well.
func New(topLevelFlags *flag.FlagSet, name string) *Commander {
	cdr := &Commander{
		topFlags: topLevelFlags,
		name:     name,
		Output:   os.Stdout,
		Error:    os.Stderr,
	}

	cdr.Explain = cdr.explain
	cdr.ExplainGroup = explainGroup
	cdr.ExplainCommand = explain
	topLevelFlags.Usage = func() { cdr.Explain(cdr.Error) }
	return cdr
}

// Name returns the commander's name
func (cdr *Commander) Name() string {
	return cdr.name
}

// Register adds a subcommand to the supported subcommands in the
// specified group. (Help output is sorted and arranged by group name.)
// The empty string is an acceptable group name; such subcommands are
// explained first before named groups.
func (cdr *Commander) Register(cmd Command, group string) {
	for _, g := range cdr.commands {
		if g.name == group {
			g.commands = append(g.commands, cmd)
			return
		}
	}
	cdr.commands = append(cdr.commands, &CommandGroup{
		name:     group,
		commands: []Command{cmd},
	})
}

// ImportantFlag marks a top-level flag as important, which means it
// will be printed out as part of the output of an ordinary "help"
// subcommand.  (All flags, important or not, are printed by the
// "flags" subcommand.)
func (cdr *Commander) ImportantFlag(name string) {
	cdr.important = append(cdr.important, name)
}

// VisitGroups visits each command group in lexicographical order, calling
// fn for each.
func (cdr *Commander) VisitGroups(fn func(*CommandGroup)) {
	sort.Sort(byGroupName(cdr.commands))
	for _, g := range cdr.commands {
		fn(g)
	}
}

// VisitCommands visits each command in registered order grouped by
// command group in lexicographical order, calling fn for each.
func (cdr *Commander) VisitCommands(fn func(*CommandGroup, Command)) {
	cdr.VisitGroups(func(g *CommandGroup) {
		for _, cmd := range g.commands {
			fn(g, cmd)
		}
	})
}

// VisitAllImportant visits the important top level flags in lexicographical
// order, calling fn for each. It visits all flags, even those not set.
func (cdr *Commander) VisitAllImportant(fn func(*flag.Flag)) {
	sort.Strings(cdr.important)
	for _, name := range cdr.important {
		f := cdr.topFlags.Lookup(name)
		if f == nil {
			panic(fmt.Sprintf("Important flag (%s) is not defined", name))
		}
		fn(f)
	}
}

// VisitAll visits the top level flags in lexicographical order, calling fn
// for each. It visits all flags, even those not set.
func (cdr *Commander) VisitAll(fn func(*flag.Flag)) {
	if cdr.topFlags != nil {
		cdr.topFlags.VisitAll(fn)
	}
}

// Execute should be called once the top-level-flags on a Commander
// have been initialized. It finds the correct subcommand and executes
// it, and returns an ExitStatus with the result. On a usage error, an
// appropriate message is printed to os.Stderr, and ExitUsageError is
// returned. The additional args are provided as-is to the Execute method
// of the selected Command.
func (cdr *Commander) Execute() ExitStatus {
	if cdr.topFlags.NArg() < 1 {
		cdr.topFlags.Usage()
		return ExitUsageError
	}

	name := cdr.topFlags.Arg(0)

	for _, group := range cdr.commands {
		for _, cmd := range group.commands {
			if name != cmd.Name() {
				continue
			}
			f := flag.NewFlagSet(name, flag.ContinueOnError)
			f.Usage = func() { cdr.ExplainCommand(cdr.Error, cmd) }
			cmd.SetFlags(f)
			if f.Parse(cdr.topFlags.Args()[1:]) != nil {
				return ExitUsageError
			}
			return cmd.Execute(f)
		}
	}

	// Cannot find this command.
	cdr.topFlags.Usage()
	return ExitUsageError
}

// countFlags returns the number of top-level flags defined, even those not set.
func (cdr *Commander) countTopFlags() int {
	count := 0
	cdr.VisitAll(func(*flag.Flag) {
		count++
	})
	return count
}

// Sorting of a slice of command groups.
type byGroupName []*CommandGroup

// TODO Sort by function rather than implement sortable?
func (p byGroupName) Len() int           { return len(p) }
func (p byGroupName) Less(i, j int) bool { return p[i].name < p[j].name }
func (p byGroupName) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

// explain prints a brief description of all the subcommands and the
// important top-level flags.
func (cdr *Commander) explain(w io.Writer) {
	fmt.Fprint(w, "Usage:\n")
	fmt.Fprintf(w, "   %s <command>\n\n", cdr.name)

	sort.Sort(byGroupName(cdr.commands))
	for _, group := range cdr.commands {
		cdr.ExplainGroup(w, group)
	}
	if cdr.topFlags == nil {
		fmt.Fprintln(w, "\nNo top level flags.")
		return
	}

	sort.Strings(cdr.important)
	if len(cdr.important) == 0 {
		if cdr.countTopFlags() > 0 {
			fmt.Fprintf(w, "\nUse \"%s flags\" for a list of top-level flags\n", cdr.name)
		}
		return
	}

	fmt.Fprintf(w, "\nTop-level flags (use \"%s flags\" for a full list):\n", cdr.name)
	for _, name := range cdr.important {
		f := cdr.topFlags.Lookup(name)
		if f == nil {
			panic(fmt.Sprintf("Important flag (%s) is not defined", name))
		}
		fmt.Fprintf(w, "  -%s=%s: %s\n", f.Name, f.DefValue, f.Usage)
	}
}

// Sorting of the commands within a group.
func (g CommandGroup) Len() int           { return len(g.commands) }
func (g CommandGroup) Less(i, j int) bool { return g.commands[i].Name() < g.commands[j].Name() }
func (g CommandGroup) Swap(i, j int)      { g.commands[i], g.commands[j] = g.commands[j], g.commands[i] }

// explainGroup explains all the commands for a particular group.
func explainGroup(w io.Writer, group *CommandGroup) {
	if len(group.commands) == 0 {
		return
	}
	if group.name == "" {
		fmt.Fprintf(w, "Commands:\n")
	} else {
		fmt.Fprintf(w, "Commands for %s:\n", group.name)
	}
	sort.Sort(group)

	tw := tabwriter.NewWriter(w, 0, 3, 3, ' ', 0)

	for _, cmd := range group.commands {
		fmt.Fprintf(tw, "   %s\t%s\n", cmd.Name(), cmd.Synopsis())
	}
	tw.Flush()
	fmt.Fprintln(w)
}

// explainCmd prints a brief description of a single command.
func explain(w io.Writer, cmd Command) {
	fmt.Fprintf(w, "Synopsis:\n")
	fmt.Fprintf(w, "   %s\n\n", cmd.Synopsis())

	fmt.Fprintf(w, "Usage:\n")
	fmt.Fprintf(w, "   %s\n\n", cmd.Usage())

	subflags := flag.NewFlagSet(cmd.Name(), flag.PanicOnError)
	subflags.SetOutput(w)
	cmd.SetFlags(subflags)
	PrintDefaults(subflags, "Flags:\n")
}

// A helper is a Command implementing a "help" command for
// a given Commander.
type helper Commander

func (h *helper) Name() string { return "help" }
func (h *helper) Synopsis() string {
	return "Show a list of all commands or describe a specific command."
}
func (h *helper) SetFlags(*flag.FlagSet) {}
func (h *helper) Usage() string {
	return `help [<command>]
	
With an argument, prints detailed information on the use of
the specified command. With no argument, print a list of
all commands and a brief description of each.
`
}
func (h *helper) Execute(f *flag.FlagSet) ExitStatus {
	switch f.NArg() {
	case 0:
		(*Commander)(h).Explain(h.Output)
		return ExitSuccess

	case 1:
		for _, group := range h.commands {
			for _, cmd := range group.commands {
				if f.Arg(0) != cmd.Name() {
					continue
				}
				(*Commander)(h).ExplainCommand(h.Output, cmd)
				return ExitSuccess
			}
		}
		fmt.Fprintf(h.Error, "Subcommand %s not understood\n", f.Arg(0))
	}

	f.Usage()
	return ExitUsageError
}

// HelpCommand returns a Command which implements a "help" subcommand.
func (cdr *Commander) HelpCommand() Command {
	return (*helper)(cdr)
}
