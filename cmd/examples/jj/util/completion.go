package util

import "fmt"

type completionOptions struct {
	Bash, Zsh, Fish bool // print a completion script for the given shell
}

// Print a command-line-completion script.
func (u *Util) Completion(opts *completionOptions) {
	fmt.Println("completion", u, opts)
}
