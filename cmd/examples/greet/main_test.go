package main

import (
	"context"
	"testing"

	"github.com/avamsi/climate"
	"github.com/avamsi/climate/testing/clitest"
	"github.com/google/go-cmp/cmp"
)

func TestMain(t *testing.T) {
	var (
		greet = clitest.New(climate.Func(greet), climate.WithMetadata(md))
		want  = `Greet someone.

Usage:
  greet [opts]

Flags:
  -g, --greeting string (default Hello)  greeting to use         
  -n, --name     string (default World)  name to greet           
  -t, --times    int                     number of times to greet
  -h, --help                             help for greet
`
	)
	// TODO(golang/go#36532): replace with t.Context().
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	got := greet(ctx, []string{"--help"}).Stdout
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("want:\n%v", want)
		t.Errorf("got:\n%v", got)
		t.Errorf("diff(-want +got):\n%v", diff)
	}
}
