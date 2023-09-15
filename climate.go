package climate

import (
	"context"
	"errors"
	"reflect"

	"github.com/avamsi/ergo/check"

	"github.com/avamsi/climate/internal"
)

type plan interface {
	execute(context.Context, *internal.Metadata) error
}

func Func(f any) *funcPlan {
	t := reflect.TypeOf(f)
	check.Truef(t.Kind() == reflect.Func, "not a func: %q", t)
	v := reflect.ValueOf(f)
	return &funcPlan{reflection{ot: t, ov: &v}}
}

var _ plan = (*funcPlan)(nil)

func Struct[T any](subcommands ...*structPlan) *structPlan {
	var (
		ptr = reflect.TypeOf((*T)(nil))
		t   = ptr.Elem()
	)
	check.Truef(t.Kind() == reflect.Struct, "not a struct: %q", t)
	return &structPlan{
		reflection{ptr: &reflection{ot: ptr}, ot: t},
		subcommands,
	}
}

var _ plan = (*structPlan)(nil)

type runOptions struct {
	metadata *[]byte
}

func Metadata(b []byte) func(*runOptions) {
	return func(opts *runOptions) {
		opts.metadata = &b
	}
}

func Run(p plan, mods ...func(*runOptions)) int {
	opts := runOptions{}
	for _, mod := range mods {
		mod(&opts)
	}
	var md *internal.Metadata
	if opts.metadata != nil {
		md = internal.DecodeAsMetadata(*opts.metadata)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// Cobra already prints the error to stderr, so just return exit code here.
	if err := p.execute(ctx, md); err != nil {
		var eerr *exitError
		if errors.As(err, &eerr) {
			return eerr.code
		}
		return 1
	}
	return 0
}
