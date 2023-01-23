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

type doc struct {
	Long, Short, Usage string
}

func (d doc) shortOrAutoGen() string {
	if d.Short != "" {
		return d.Short
	}
	s := d.Long
	i := strings.Index(s, "\n\n")
	if i != -1 {
		s = s[:i]
	}
	s = strings.Join(strings.Fields(s), " ")
	if s == "" {
		return ""
	}
	rs := []rune(s)
	rs[0] = unicode.ToUpper(rs[0])
	if len(rs) > 80 {
		rs = append(rs[:77], []rune("...")...)
	} else if len(rs) > 1 && rs[len(rs)-1] == '.' {
		// Clip the period at the end by convention but only if the last but one
		// character is a letter or a digit. TODO: other cases.
		if unicode.IsLetter(rs[len(rs)-2]) || unicode.IsDigit(rs[len(rs)-2]) {
			rs = rs[:len(rs)-1]
		}
	}
	return string(rs)
}

func (d doc) usageOr(s string) string {
	if d.Usage != "" {
		return d.Usage
	}
	return s
}

var docs map[string]doc

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
