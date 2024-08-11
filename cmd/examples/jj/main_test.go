package main

import (
	"context"
	"testing"

	"github.com/avamsi/climate"
	"github.com/avamsi/climate/cmd/examples/jj/util"
	"github.com/avamsi/climate/testing/clitest"
	"github.com/google/go-cmp/cmp"
)

func TestMain(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "jj--help",
			args: []string{"--help"},
			want: `Jujutsu (an experimental VCS).

Usage:
  jj [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  git         Commands for working with the underlying Git repo
  help        Help about any command
  init        Create a new repo in the given directory
  squash      Move changes from a revision into its parent
  util        Infrequently used commands such as for generating shell completions

Flags:
  -R, --repository          path  path to the repo to operate on          
      --ignore-working-copy       don't snapshot / update the working copy
  -h, --help                      help for jj

Use "jj [command] --help" for more information about a command.
`,
		},
		{
			name: "jj-git--help",
			args: []string{"git", "--help"},
			want: `Commands for working with the underlying Git repo.

Usage:
  jj git [command]

Available Commands:
  export      Update the underlying Git repo with changes made in the repo
  remote      Manage Git remotes

Flags:
  -h, --help  help for git

Global Flags:
      --ignore-working-copy       don't snapshot / update the working copy
  -R, --repository          path  path to the repo to operate on

Use "jj git [command] --help" for more information about a command.
`,
		},
	}
	var (
		p  = climate.Struct[jj](climate.Struct[git](), climate.Struct[util.Util]())
		jj = clitest.New(p, climate.WithMetadata(md))
	)
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// TODO(golang/go#36532): replace with t.Context().
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			got := jj(ctx, test.args).Stdout
			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("want:\n%v", test.want)
				t.Errorf("got:\n%v", got)
				t.Errorf("diff(-want +got):\n%v", diff)
			}
		})
	}
}
