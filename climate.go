package climate

import (
	"context"
	"reflect"

	"github.com/avamsi/climate/internal"

	"github.com/avamsi/ergo/check"
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
		md = internal.DecodeMetadata(*opts.metadata)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// We already print the error to stderr, so just return exit code here.
	if p.execute(ctx, md) != nil {
		return 1
	}
	return 0
}
