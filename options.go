package eclipse

import (
	"log"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/avamsi/ergo"

	flag "github.com/spf13/pflag"
)

var (
	anyUpperLower = regexp.MustCompile("(.)([A-Z][a-z])")
	lowerUpper    = regexp.MustCompile("([a-z])([A-Z])")
)

func toKebabCase(s string) string {
	s = anyUpperLower.ReplaceAllString(s, "${1}-${2}")
	s = lowerUpper.ReplaceAllString(s, "${1}-${2}")
	return strings.ToLower(s)
}

type options struct {
	t           reflect.Type
	v           reflect.Value
	parentIndex int
}

func newOptions(t reflect.Type, fs *flag.FlagSet, parentID string) *options {
	if t.Kind() != reflect.Struct {
		log.Fatalf("want: struct; got: '%#v'", t)
	}
	opts := &options{
		t:           t,
		v:           reflect.New(t).Elem(),
		parentIndex: -1,
	}
	opts.declareFlags(fs, parentID)
	return opts
}

func flagVar[T bool | int64 | string](
	flagTVar func(*T, string, T, string),
	fv reflect.Value,
	sf reflect.StructField,
	strconvF func(string) (T, error),
	parentID string,
) {
	defaultTag, ok := sf.Tag.Lookup("default")
	var defaultValue T
	if ok {
		defaultValue = ergo.Check1(strconvF(defaultTag))
	}
	ptr := (*T)(fv.Addr().UnsafePointer())
	// TODO: consider adding support for shorthand flags.
	flagTVar(ptr, toKebabCase(sf.Name), defaultValue, docs[parentID+"."+sf.Name])
}

func (opts *options) declareFlags(fs *flag.FlagSet, parentID string) {
	for i := 0; i < opts.t.NumField(); i++ {
		sf, fv := opts.t.Field(i), opts.v.Field(i)
		switch sf.Type.Kind() {
		case reflect.Bool:
			flagVar(fs.BoolVar, fv, sf, strconv.ParseBool, parentID)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			s2i64 := func(s string) (int64, error) {
				return strconv.ParseInt(s, 10, 64)
			}
			flagVar(fs.Int64Var, fv, sf, s2i64, parentID)
		case reflect.String:
			s2s := func(s string) (string, error) { return s, nil }
			flagVar(fs.StringVar, fv, sf, s2s, parentID)
		case reflect.Struct:
			if opts.parentIndex != -1 {
				log.Fatalf("want: exactly one struct field; got: '%#v'", opts.t)
			}
			opts.parentIndex = i
		default:
			log.Fatalf("want: bool|int|string fields; got: '%#v'", sf)
		}
	}
}
