# Climate

Climate "CLI Mate" aims to make creating CLIs in Go easy (and fun!), similar to
[python-fire](https://github.com/google/python-fire).  
It's also built on top of [Cobra](https://github.com/spf13/cobra) and so comes
with "batteries included" (help, shell completion etc.).

## Usage

https://github.com/avamsi/climate/blob/541bffe22424d1d83d1044ca4f1f0629f4588b76/cmd/examples/greet/main.go#L1-L55

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

https://github.com/avamsi/climate/blob/541bffe22424d1d83d1044ca4f1f0629f4588b76/cmd/examples/jj/main.go#L1-L89

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
