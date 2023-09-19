package climate

import (
	"context"
	"os"
	"os/exec"
	"reflect"

	"github.com/avamsi/ergo/assert"

	"github.com/avamsi/climate/internal"
)

type plan interface {
	execute(context.Context, *internal.Metadata) error
}

func Func(f any) *funcPlan {
	t := reflect.TypeOf(f)
	assert.Truef(t.Kind() == reflect.Func, "not a func: %q", t)
	v := reflect.ValueOf(f)
	return &funcPlan{reflection{ot: t, ov: &v}}
}

var _ plan = (*funcPlan)(nil)

func Struct[T any](subcommands ...*structPlan) *structPlan {
	var (
		ptr = reflect.TypeOf((*T)(nil))
		t   = ptr.Elem()
	)
	assert.Truef(t.Kind() == reflect.Struct, "not a struct: %q", t)
	return &structPlan{
		reflection{ptr: &reflection{ot: ptr}, ot: t},
		subcommands,
	}
}

var _ plan = (*structPlan)(nil)

func exitCode(err error) int {
	if err == nil { // if _no_ error
		return 0
	}
	switch err := err.(type) {
	case *exitError:
		return err.code
	case *exec.ExitError:
		return err.ExitCode()
	default:
		return 1
	}
}

type runOptions struct {
	metadata *[]byte
}

func WithMetadata(b []byte) func(*runOptions) {
	return func(opts *runOptions) {
		opts.metadata = &b
	}
}

func Run(ctx context.Context, p plan, mods ...func(*runOptions)) int {
	var opts runOptions
	for _, mod := range mods {
		mod(&opts)
	}
	var md *internal.Metadata
	if opts.metadata != nil {
		md = internal.DecodeAsMetadata(*opts.metadata)
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	// Cobra already prints the error to stderr, so just return exit code here.
	return exitCode(p.execute(ctx, md))
}

func RunAndExit(p plan, mods ...func(*runOptions)) {
	os.Exit(Run(context.Background(), p, mods...))
}
