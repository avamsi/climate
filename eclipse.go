package eclipse

import (
	"bytes"
	"encoding/gob"
	"log"
	"reflect"
	"strings"

	"github.com/avamsi/ergo"
	"github.com/spf13/cobra"
)

var docs map[string]string

func parsedCmdDocFor(s string) (long, short, usage string) {
	lines := []string{}
	for _, l := range strings.Split(docs[s], "\n") {
		if strings.HasPrefix(l, "Short: ") {
			if short != "" {
				panic("TODO")
			}
			short = strings.TrimPrefix(l, "Short: ")
		} else if strings.HasPrefix(l, "Usage: ") {
			if usage != "" {
				panic("TODO")
			}
			usage = strings.TrimPrefix(l, "Usage: ")
		} else {
			lines = append(lines, l)
		}
	}
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
	return strings.Join(lines, "\n"), short, usage
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
		log.Fatalf("want: exactly one root command; got: '%#v'", rootCobraCmds)
	}
	for rootCobraCmd := range rootCobraCmds {
		rootCobraCmd.Execute()
	}
}
