package eclipse

import (
	"encoding/base64"
	"encoding/json"
	"log"
	"reflect"
	"strings"

	"github.com/avamsi/ergo"
	"github.com/spf13/cobra"
)

var docs map[string]string

func longAndShortDocsFor(s string) (long, short string) {
	long, short = docs[s], docs[s]
	i := strings.Index(long, "\n\n")
	if i != -1 {
		short = long[:i]
	}
	return long, short
}

func Execute(args ...any) {
	if rawDocs, ok := args[0].(string); ok {
		ergo.Check0(json.Unmarshal(ergo.Check1(base64.StdEncoding.DecodeString(rawDocs)), &docs))
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
