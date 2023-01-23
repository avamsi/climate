package main

import (
	"fmt"

	_ "embed"

	"github.com/avamsi/clifr"
)

// Cobra is a CLI library for Go that empowers applications.
// This application is a tool to generate the needed files
// to quickly create a Cobra application.
//
//clifr:usage cobra [command]
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
	// list of flags for this command
	Flags []string
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
//clifr:short Add a command to a Cobra Application
//clifr:usage add [command name]
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
//clifr:short Initialize a Cobra Application
//clifr:usage init [path] [flags]
func (p Pretend) Init() {
	fmt.Println("cobra pretend init", p)
}

//go:generate clifr docs --out=docs.clifr
//go:embed docs.clifr
var docs []byte

func main() {
	clifr.Execute(docs, Cobra{}, Pretend{})
}
