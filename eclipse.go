package eclipse

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/avamsi/ergo"
	"github.com/spf13/cobra"
)

var docs map[string]string

const (
	shortDirective = "Short: "
	usageDirective = "Usage: "
)

func parsedCmdDocFor(s string) (long, short, usage string) {
	lines := []string{}
	for _, l := range strings.Split(docs[s], "\n") {
		if strings.HasPrefix(l, shortDirective) {
			if short != "" {
				fmt.Fprintf(os.Stderr, "got: '%#v'; want: exactly one short directive\n", l)
				os.Exit(1)
			}
			short = l[len(shortDirective):]
		} else if strings.HasPrefix(l, "Usage: ") {
			if usage != "" {
				fmt.Fprintf(os.Stderr, "got: '%#v'; want: exactly one usage directive\n", l)
				os.Exit(1)
			}
			usage = l[len(usageDirective):]
		} else {
			lines = append(lines, l)
		}
	}
	long = strings.Join(lines, "\n")
	if short == "" {
		i := strings.Index(long, "\n\n")
		if i != -1 {
			short = long[:i]
		}
		short = strings.Join(strings.Fields(short), " ")
		if len(short) > 80 {
			short = short[:77] + "..."
		}
	}
	return long, short, usage
}

func Execute(args ...any) {
	if rawDocs, ok := args[0].([]byte); ok {
		ergo.Must0(gob.NewDecoder(bytes.NewBuffer(rawDocs)).Decode(&docs))
		args = args[1:]
	}
	cobraCmds := []*cobra.Command{}
	for _, arg := range args {
		cmd := typeToCommand(reflect.TypeOf(arg))
		cobraCmds = append(cobraCmds, cmd.cobraCmd)
	}
	rootCobraCmds := map[*cobra.Command]bool{}
	for _, cobraCmd := range cobraCmds {
		rootCobraCmds[cobraCmd.Root()] = true
	}
	if len(rootCobraCmds) != 1 {
		fmt.Fprintf(os.Stderr, "got: '%#v'; want: exactly one root command\n", rootCobraCmds)
		os.Exit(1)
	}
	for rootCobraCmd := range rootCobraCmds {
		rootCobraCmd.Execute()
	}
}
