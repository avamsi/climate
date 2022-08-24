# Eclipse

Eclipse aims to make creating CLIs in Go easy (and fun!), similar to [python-fire](https://github.com/google/python-fire).

## Usage

If you want to skip the incremental introduction and would instead prefer seeing working code for a full feature demo,
jump to the [tl;dr](#tldr) section below.

### Commands

With Eclipse, to create a command, all you need to do is create a struct (and then pass it to `eclipse.Execute`).
To make the "root" command itself runnable, you just add an `Execute` (note that it's exported) method to the struct.
And to add subcommands, you just add more exported methods to the struct --

```go
type Cobra struct{}

func (Cobra) Execute() {
	fmt.Println("cobra")
}

func (Cobra) Init() {
	fmt.Println("cobra init")
}

func (Cobra) Add() {
	fmt.Println("cobra add")
}
```

Eclipse is built on top of [Cobra](https://github.com/spf13/cobra) (and through it), autogenerates completion and help commands for you --

```
$ cobra --help

Usage:
  cobra [flags]
  cobra [command]

Available Commands:
  add
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  init

Flags:
  -h, --help   help for command

Use "cobra [command] --help" for more information about a command.
```

### Flags

Eclipse also makes adding flags easy.
To add "global" flags (i.e., something that's applicable for all the commands i.e., the "root" command and all subcommands),
you just add exported value fields to the struct.

To add "local" flags (i.e., something that's only applicable for a single command),
you just create a (potentially anonymous) struct (again, with exported value fields) and add it as a param to the respective method.

This would look something like --

```go
type Cobra struct {
	Author  string
	Config  string
	License string
}

func (c Cobra) Init() {
	fmt.Println("cobra init", c)
}

type AddOpts struct {
	Parent string
}

func (c Cobra) Add(opts AddOpts) {
	fmt.Println("cobra add", c, opts)
}
```

Let's also see what the help page looks like --

```
$ cobra --help

Usage:
  cobra [flags]
  cobra [command]

Available Commands:
  add
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  init

Flags:
      --author string
      --config string
  -h, --help             help for cobra
      --license string

Use "cobra [command] --help" for more information about a command.
```

And for a subcommand --

```
$ cobra add --help

Usage:
  cobra add [flags]

Flags:
  -h, --help            help for add
      --parent string

Global Flags:
      --author string
      --config string
      --license string
```

### Subcommand-ception

To create more "parent" commands (i.e., commands with subcommands, like the root command),
you just make them a struct with the root command or some other parent command as one of the fields. For example --

```go
type Cobra struct {
	Author  string
	Config  string
	License string
}

type Pretend struct {
	c Cobra
}

func (p Pretend) Init() {
	fmt.Println("cobra pretend init", p)
}
```

### Default values, shorthands and docs

Eclipse also supports setting default values and shorthands for flags through respective struct tags and
is able to autogenerate help documentation from Godocs if you're willing to suffer an extra step
(you'll need to install the Eclipse CLI and run `go generate ./...` on any Godoc changes).

``` shell
$ go install github.com/avamsi/eclipse/cli/eclipse@latest
$ go generate ./...
```

### tl;dr

Here's the code for a full feature demo --

```go
package main

import (
	"fmt"

	_ "embed"

	"github.com/avamsi/eclipse"
)

// Cobra is a CLI library for Go that empowers applications.
// This application is a tool to generate the needed files
// to quickly create a Cobra application.
//
// Usage: cobra [command]
type Cobra struct {
	// author name for copyright attribution
	Author string `short:"a" default:"YOUR NAME"`
	Config string `default:"$HOME/.cobra.yaml"` // config file
	// name of license for the project
	License string `short:"l"`
}

type AddOpts struct {
	// variable name of parent command for this command
	Parent string `short:"p" default:"rootCmd"`
}

// Add (cobra add) will create a new command, with a license and
// the appropriate structure for a Cobra-based CLI application,
// and register it to its parent (default rootCmd).
//
// If you want your command to be public, pass in the command name
// with an initial uppercase letter.
//
// Example: cobra add server -> resulting in a new cmd/server.go
//
// Short: Add a command to a Cobra Application
// Usage: add [command name]
func (c Cobra) Add(opts AddOpts, args []string) {
	fmt.Println("cobra add", c, opts, args)
}

type Pretend struct {
	C Cobra
}

// Initialize (cobra init) will create a new application, with a license
// and the appropriate structure for a Cobra-based CLI application.
//
// Cobra init must be run inside of a go module (please run "go mod init <MODNAME>" first)
//
// Short: Initialize a Cobra Application
// Usage: init [path] [flags]
func (p Pretend) Init() {
	fmt.Println("cobra pretend init", p)
}

//go:generate eclipse docs --out=eclipse.docs
//go:embed eclipse.docs
var docs []byte

func main() {
	eclipse.Execute(docs, Cobra{}, Pretend{})
}
```

Let's also see what the help page looks like --

```
$ cobra --help

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.

Usage:
  cobra [command]

Available Commands:
  add         Add a command to a Cobra Application
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  pretend

Flags:
  -a, --author string    author name for copyright attribution (default "YOUR NAME")
      --config string    config file (default "$HOME/.cobra.yaml")
  -h, --help             help for cobra
  -l, --license string   name of license for the project

Use "cobra [command] --help" for more information about a command.
```

And for a subcommand --

```
$ cobra add --help

Add (cobra add) will create a new command, with a license and
the appropriate structure for a Cobra-based CLI application,
and register it to its parent (default rootCmd).

If you want your command to be public, pass in the command name
with an initial uppercase letter.

Example: cobra add server -> resulting in a new cmd/server.go

Usage:
  cobra add [command name] [flags]

Flags:
  -h, --help            help for add
  -p, --parent string   variable name of parent command for this command (default "rootCmd")

Global Flags:
  -a, --author string    author name for copyright attribution (default "YOUR NAME")
      --config string    config file (default "$HOME/.cobra.yaml")
  -l, --license string   name of license for the project
```
