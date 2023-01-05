package clifr

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"os"
	"reflect"
	"strings"
	"unicode"

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
		short = long
		i := strings.Index(short, "\n\n")
		if i != -1 {
			short = short[:i]
		}
		// TODO: another possibility here instead of trying to fit this on one
		// line is to indent the following lines so they're still readable.
		short = strings.Join(strings.Fields(short), " ")
		if short != "" {
			runes := []rune(short)
			runes[0] = unicode.ToUpper(runes[0])
			if len(runes) > 80 {
				runes = append(runes[:77], []rune("...")...)
			} else if len(runes) > 1 && runes[len(runes)-1] == '.' {
				// Clip the period at the end by convention but only if the
				// last but one character is a letter or a digit.
				// TODO: do something about the other cases?
				if unicode.IsLetter(runes[len(runes)-2]) || unicode.IsDigit(runes[len(runes)-2]) {
					runes = runes[:len(runes)-1]
				}
			}
			short = string(runes)
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
