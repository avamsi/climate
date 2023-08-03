package climate

import (
	"reflect"
	"strings"
	"unsafe"

	"github.com/avamsi/ergo/check"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"golang.org/x/exp/utf8string"

	"github.com/avamsi/climate/internal"
)

type flagTypeVarP[T any] func(*T, string, string, T, string)

type tags struct {
	m map[string]string
}

func newTags(st reflect.StructTag) tags {
	m := make(map[string]string)
	if v, ok := st.Lookup("default"); ok {
		m["default"] = v
	}
	for _, kv := range strings.Split(st.Get("climate"), ",") {
		k, v, _ := strings.Cut(kv, "=")
		m[k] = v
	}
	return tags{m}
}

func (ts tags) shorthand() string {
	return ts.m["short"]
}

func (ts tags) defaultValue() (string, bool) {
	v, ok := ts.m["default"]
	return v, ok
}

func (ts tags) required() bool {
	_, ok := ts.m["required"]
	return ok
}

type option struct {
	fset *pflag.FlagSet
	t    reflect.Type
	p    unsafe.Pointer
	name string
	tags
	usage string
}

func declareOption[T any](flagVarP flagTypeVarP[T], opt *option, typer typeParser[T]) {
	var (
		p     = (*T)(opt.p)
		value T
	)
	if v, ok := opt.defaultValue(); ok {
		value = typer(v)
	}
	check.Truef(utf8string.NewString(opt.name).IsASCII(), "not ASCII: %q", opt.name)
	shorthand := opt.shorthand()
	if shorthand == "" {
		shorthand = strings.ToLower(opt.name[:1])
	}
	flagVarP(p, opt.name, shorthand, value, opt.usage)
	if opt.required() {
		check.Nil(cobra.MarkFlagRequired(opt.fset, opt.name))
	}
}

func (opt *option) declare() bool {
	switch k := opt.t.Kind(); k {
	case reflect.Bool:
		declareOption(
			opt.fset.BoolVarP,
			opt,
			parseBool,
		)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		declareOption(
			opt.fset.Int64VarP,
			opt,
			parseInt64,
		)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		declareOption(
			opt.fset.Uint64VarP,
			opt,
			parseUint64,
		)
	case reflect.Float32, reflect.Float64:
		declareOption(
			opt.fset.Float64VarP,
			opt,
			parseFloat64,
		)
	case reflect.String:
		declareOption(
			opt.fset.StringVarP,
			opt,
			parseString,
		)
	case reflect.Slice:
		switch e := opt.t.Elem(); e.Kind() {
		case reflect.Bool:
			declareOption(
				opt.fset.BoolSliceVarP,
				opt,
				sliceParser(parseBool),
			)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			declareOption(
				opt.fset.Int64SliceVarP,
				opt,
				sliceParser(parseInt64),
			)
		case reflect.Float32, reflect.Float64:
			declareOption(
				opt.fset.Float64SliceVarP,
				opt,
				sliceParser(parseFloat64),
			)
		case reflect.String:
			declareOption(
				opt.fset.StringSliceVarP,
				opt,
				sliceParser(parseString),
			)
		default:
			internal.Panicf("not []bool | []Signed | []Float | []string: %q", e)
		}
	default:
		if typeIsStructPointer(opt.t) {
			return false
		}
		internal.Panicf("not bool | Integer | Float | string | []T: %q", opt.t)
	}
	return true
}

type options struct {
	reflection
	parent *reflection
	fset   *pflag.FlagSet
	md     *internal.Metadata
}

func (opts *options) declare() {
	parentSet := (opts.parent == nil)
	for i := 0; i < opts.t().NumField(); i++ {
		var (
			f  = opts.t().Field(i)
			md = opts.md.Child(f.Name)
		)
		// Long() returns the "Doc" part of the field and Short() returns the
		// "Comment" part. Other Metadata is neither collected, nor used.
		usage := md.Long()
		if usage == "" {
			usage = md.Short()
		}
		var (
			v   = opts.v().Field(i)
			opt = option{
				fset:  opts.fset,
				t:     f.Type,
				p:     v.Addr().UnsafePointer(),
				name:  f.Name,
				tags:  newTags(f.Tag),
				usage: usage,
			}
		)
		if !opt.declare() {
			if opts.parent == nil {
				internal.Panicf("not bool | Integer | Float | string | []T: %q", f.Type)
			}
			if f.Type != opts.parent.ptr.t() {
				internal.Panicf(
					"not bool | Integer | Float | string | []T | %q: %q",
					opts.parent.t(), f.Type)
			}
			if parentSet {
				internal.Panicf("more than one parent: %q", f.Type)
			}
			v.Set(*opts.parent.ptr.v())
		}
	}
}
