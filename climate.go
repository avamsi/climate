// Package climate "CLI Mate" provides a set of APIs to autogenerate CLIs from
// structs/functions with support for nested subcommands, global/local flags,
// help generation from comments, typo suggestions, shell completion and more.
//
// See https://github.com/avamsi/climate/blob/main/README.md for more details.
package climate

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"reflect"

	"github.com/avamsi/ergo"
	"github.com/avamsi/ergo/assert"

	"github.com/avamsi/climate/internal"
)

// Func returns an executable plan for the given function, which must conform to
// the following signatures (excuse the partial [optional] notation):
//
//	func([ctx context.Context], [opts *T], [args []string]) [(err error)]
//
// All of ctx, opts, args and error are optional. If opts is present, T must be
// a struct (whose fields are used as flags).
func Func(f any) *funcPlan {
	t := reflect.TypeOf(f)
	assert.Truef(t.Kind() == reflect.Func, "not a func: %v", t)
	v := reflect.ValueOf(f)
	return &funcPlan{reflection{ot: t, ov: &v}}
}

var _ internal.Plan = (*funcPlan)(nil)

// Struct returns an executable plan for the struct given as the type parameter,
// with its methods* (and "child" structs) as subcommands.
//
// * Only methods with pointer receiver are considered (and they must otherwise
// conform to the same signatures described in Func).
func Struct[T any](subcommands ...*structPlan) *structPlan {
	t := reflect.TypeFor[T]()
	assert.Truef(t.Kind() == reflect.Struct, "not a struct: %v", t)
	if n := t.NumMethod(); n > 0 {
		ms := make([]string, n)
		for i := 0; i < n; i++ {
			ms[i] = t.Method(i).Name
		}
		ergo.Panicf("nonzero methods %v on: %v", ms, t)
	}
	ptr := reflect.PointerTo(t)
	assert.Truef(ptr.NumMethod() > 0, "no methods on: %v", ptr)
	return &structPlan{
		reflection{ptr: &reflection{ot: ptr}, ot: t},
		subcommands,
	}
}

var _ internal.Plan = (*structPlan)(nil)

func exitCode(err error) int {
	if err == nil { // if _no_ error
		return 0
	}
	if eerr := new(exitError); errors.As(err, &eerr) {
		return eerr.code
	} else if eerr := new(exec.ExitError); errors.As(err, &eerr) {
		return eerr.ExitCode()
	}
	return 1
}

// WithMetadata returns a modifier that sets the metadata to be used by Run for
// augmenting the CLI with additional information (for --help etc.).
func WithMetadata(b []byte) func(*internal.RunOptions) {
	return func(opts *internal.RunOptions) {
		opts.Metadata = &b
	}
}

// Run executes the given plan and returns the exit code.
func Run(ctx context.Context, p internal.Plan, mods ...func(*internal.RunOptions)) int {
	var opts internal.RunOptions
	for _, mod := range mods {
		mod(&opts)
	}
	var md *internal.Metadata
	if opts.Metadata != nil {
		md = internal.DecodeAsMetadata(*opts.Metadata)
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	// Cobra already prints the error to stderr, so just return exit code here.
	return exitCode(p.Execute(ctx, md))
}

// RunAndExit executes the given plan and exits with the exit code.
func RunAndExit(p internal.Plan, mods ...func(*internal.RunOptions)) {
	os.Exit(Run(context.Background(), p, mods...))
}
