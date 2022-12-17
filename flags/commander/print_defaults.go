package commander

import (
	"flag"
	"fmt"
	"strings"
	"text/tabwriter"
)

// PrintDefaults prints, to standard error unless configured otherwise, the
// default values of all defined command-line flags in the set. See the
// documentation for the global function PrintDefaults for more information.
func PrintDefaults(fs *flag.FlagSet, hdr string) {
	if countFlags(fs) > 0 {
		fmt.Fprint(fs.Output(), hdr)

		tw := tabwriter.NewWriter(fs.Output(), 0, 3, 3, ' ', 0)
		fs.VisitAll(func(f *flag.Flag) {
			typ, desc := unquoteUsage(f)
			fmt.Fprintf(tw, "  -%s %s\t%s\n", f.Name, typ, desc)
		})
		tw.Flush()
	}
}

func countFlags(fs *flag.FlagSet) (n int) {
	fs.VisitAll(func(*flag.Flag) { n++ })
	return n
}

// unquoteUsage extracts a back-quoted name from the usage
// string for a flag and returns it and the un-quoted usage.
// Given "a `name` to show" it returns ("name", "a name to show").
// If there are no back quotes, the name is an educated guess of the
// type of the flag's value, or the empty string if the flag is boolean.
func unquoteUsage(flag *flag.Flag) (name string, usage string) {
	// Look for a back-quoted name, but avoid the strings package.
	usage = flag.Usage
	for i := 0; i < len(usage); i++ {
		if usage[i] == '`' {
			for j := i + 1; j < len(usage); j++ {
				if usage[j] == '`' {
					name = usage[i+1 : j]
					usage = strings.TrimSpace(usage[j+1:]) // usage[:i] + name + usage[j+1:]
					return name, usage
				}
			}
			break // Only one back quote; use type name.
		}
	}

	return
}
