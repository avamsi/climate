# Climate

Climate "CLI Mate" aims to make creating CLIs in Go easy (and fun!), similar to
[python-fire](https://github.com/google/python-fire). It's also built on top of
[Cobra](https://github.com/spf13/cobra), so comes with "batteries included".

## Usage

https://github.com/avamsi/climate/blob/main/cmd/examples/greet/main.go

```
$ greet --help

Greet someone.

Usage:
  greet [opts]

Flags:
  -g, --greeting string   greeting to use (default "Hello")
  -h, --help              help for greet
  -n, --name string       name to greet (default "World")
```

### Subcommands

https://github.com/avamsi/climate/blob/main/cmd/examples/jj/main.go

```
$ jj --help

Jujutsu (an experimental VCS).

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
  -h, --help                  help for jj
      --ignore-working-copy   don't snapshot / update the working copy
  -R, --repository path       path to the repo to operate on

Use "jj [command] --help" for more information about a command.
```

```
$ jj git --help

Commands for working with the underlying Git repo.

Usage:
  jj git [command]

Available Commands:
  export      Update the underlying Git repo with changes made in the repo
  remote      Manage Git remotes

Flags:
  -h, --help   help for git

Global Flags:
      --ignore-working-copy   don't snapshot / update the working copy
  -R, --repository path       path to the repo to operate on

Use "jj git [command] --help" for more information about a command.
```
