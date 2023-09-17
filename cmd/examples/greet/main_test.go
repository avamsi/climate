package main

import (
	"os/exec"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestMain(t *testing.T) {
	var (
		help = exec.Command("go", "run", ".", "--help")
		want = `Greet someone.

Usage:
  greet [opts]

Flags:
  -g, --greeting string (default Hello)  greeting to use         
  -n, --name     string (default World)  name to greet           
  -t, --times    int                     number of times to greet
  -h, --help                             help for greet
`
		got, err = help.CombinedOutput()
	)
	if err != nil {
		t.Errorf("error: %v\n", err)
	}
	if diff := cmp.Diff(want, string(got)); diff != "" {
		t.Errorf("want:\n%s", want)
		t.Errorf("got:\n%s", got)
		t.Errorf("diff(-want +got):%s", diff)
	}
}
