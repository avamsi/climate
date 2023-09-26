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
		out, err = help.CombinedOutput()
	)
	if err != nil {
		t.Errorf("error: %v\n", err)
	}
	got := string(out)
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("want:\n%v", want)
		t.Errorf("got:\n%v", got)
		t.Errorf("diff(-want +got):%v", diff)
	}
}
