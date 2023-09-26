package main

import (
	"os/exec"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestMain(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "jj --help",
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
			name: "jj git --help",
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
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cmd := exec.Command("go", "run", ".")
			cmd.Args = append(cmd.Args, tc.args...)
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Errorf("error: %v\n", err)
			}
			got := string(out)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("want:\n%v", tc.want)
				t.Errorf("got:\n%v", got)
				t.Errorf("diff(-want +got):%v", diff)
			}
		})
	}
}
