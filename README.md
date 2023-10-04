# Climate

Climate "CLI Mate" aims to make creating CLIs in Go easy (and fun!), similar to
[python-fire](https://github.com/google/python-fire).  
It's also built on top of [Cobra](https://github.com/spf13/cobra) and so comes
with "batteries included" (help, shell completion etc.).

## Usage

https://github.com/avamsi/climate/blob/0f0a8617373f1f55f08797fe8c8807902f5582bd/cmd/examples/greet/main.go#L11-L54

```
$ greet --help

Greet someone.

Usage:
  greet [opts]

Flags:
  -g, --greeting string (default Hello)  greeting to use
  -n, --name     string (default World)  name to greet
  -t, --times    int                     number of times to greet
  -h, --help                             help for greet
  -v, --version                          version for greet
```

### Subcommands

https://github.com/avamsi/climate/blob/0f0a8617373f1f55f08797fe8c8807902f5582bd/cmd/examples/jj/main.go#L14-L87

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
  version     Display jj's version information

Flags:
  -R, --repository          path  path to the repo to operate on
      --ignore-working-copy       don't snapshot / update the working copy
  -h, --help                      help for jj
  -v, --version                   version for jj

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
  -h, --help  help for git

Global Flags:
      --ignore-working-copy       don't snapshot / update the working copy
  -R, --repository          path  path to the repo to operate on

Use "jj git [command] --help" for more information about a command.
```
