package main

import (
	"errors"
	"fmt"

	_ "embed"

	"github.com/avamsi/climate"
	"github.com/avamsi/climate/cmd/examples/jj/util"
)

// See ../greet/main.go first for some details that are not covered here.

// Struct is automatically converted to a command --
//	1. Struct names are converted to lowercase and used as the command name.
//	2. Struct fields are automatically declared as "global" flags.
//	3. Struct methods are automatically converted to subcommands --
//	   1. Method names are converted to lowercase and used as the command name.
//	   2. Method docs are truncated and are used* as short help strings.
//	   3. Method directives are used* to declare aliases or explicitly set the
//	      short help strings (//climate:aliases, for example).
//	4. "Sub-structs" are automatically converted to subcommands, recursively.

// Jujutsu (an experimental VCS).
type jj struct {
	Repository        string `short:"R"` // `path` to the repo to operate on
	IgnoreWorkingCopy bool   // don't snapshot / update the working copy
}

// Create a new repo in the given directory.
func (j *jj) Init(dir *string) {
	fmt.Println("init", j, dir)
}

type squashOptions struct {
	Revision    string `short:"r" default:"@"`
	Interactive bool   `short:"i"` // interactively choose which parts to squash
}

// Move changes from a revision into its parent.
//
// After moving the changes into the parent, the child revision will have the
// same content state as before. If that means that the change is now empty
// compared to its parent, it will be abandoned. Without `--interactive`, the
// child change will always be empty.
//
//climate:aliases am, amend
func (j *jj) Squash(opts *squashOptions, paths [5]string) {
	fmt.Println("squash", j, opts, paths)
}

// Commands for working with the underlying Git repo.
type git struct {
	J *jj
}

// Manage Git remotes.
func (g *git) Remote() error {
	return errors.New("not implemented")
}

// Update the underlying Git repo with changes made in the repo.
func (g *git) Export() {
	fmt.Println("export", g.J)
}

//go:generate go run github.com/avamsi/climate/cmd/climate --out=md.climate
//go:embed md.climate
var md []byte

func main() {
	// Note the recursive struct embedding below, which lets us create "deep"
	// subcommands like this (indentation implies subcommand) --
	//
	// jj
	//     init
	//     squash
	//     git
	//         remote
	//         export
	//     util
	//         completion
	climate.Run(
		climate.Struct[jj](climate.Struct[git](), climate.Struct[util.Util]()),
		climate.Metadata(md))
}
