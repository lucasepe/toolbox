package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/lucasepe/toolbox/flags/commander"
)

const (
	appName = "dummy"
)

func main() {

	app := commander.New(flag.CommandLine, appName)
	app.Register(app.HelpCommand(), "")
	app.Register(&printCmd{}, "")

	flag.Parse()

	os.Exit(int(app.Execute()))
}

type printCmd struct {
	capitalize bool
}

func (*printCmd) Name() string     { return "print" }
func (*printCmd) Synopsis() string { return "Print args to stdout." }
func (*printCmd) Usage() string    { return "print [-capitalize] <some text>" }

func (p *printCmd) SetFlags(f *flag.FlagSet) {
	f.BoolVar(&p.capitalize, "capitalize", false, "capitalize output")
}

func (p *printCmd) Execute(f *flag.FlagSet) commander.ExitStatus {
	for _, arg := range f.Args() {
		if p.capitalize {
			arg = strings.ToUpper(arg)
		}
		fmt.Printf("%s ", arg)
	}
	fmt.Println()
	return commander.ExitSuccess
}
