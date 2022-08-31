package eclipse

import (
	"encoding/csv"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"unsafe"

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
		fmt.Fprintf(os.Stderr, "got: '%#v'; want: struct\n", t)
		os.Exit(1)
	}
	opts := &options{
		t:           t,
		v:           reflect.New(t).Elem(),
		parentIndex: -1,
	}
	opts.declareFlags(fs, parentID)
	return opts
}

type flagVarOpts[T any] struct {
	flagTVar func(*T, string, string, T, string)
	ptr      unsafe.Pointer
	sf       reflect.StructField
	s2t      func(string) (T, error)
	usage    string
}

func flagVar[T any](opts flagVarOpts[T]) {
	name := toKebabCase(opts.sf.Name)
	shorthand := opts.sf.Tag.Get("short")
	defaultTag, ok := opts.sf.Tag.Lookup("default")
	var defaultValue T
	if ok {
		defaultValue = ergo.Must1(opts.s2t(defaultTag))
	}
	opts.flagTVar((*T)(opts.ptr), name, shorthand, defaultValue, opts.usage)
}

func s2i64(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

func s2u64(s string) (uint64, error) {
	return strconv.ParseUint(s, 10, 64)
}

func s2f64(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}

func s2s(s string) (string, error) {
	return s, nil
}

func toSlice[T any](s2t func(string) (T, error)) func(string) ([]T, error) {
	return func(s string) ([]T, error) {
		if s == "" {
			return []T{}, nil
		}
		ts := []T{}
		for _, ss := range strings.Split(s, ",") {
			t, err := s2t(ss)
			if err != nil {
				return nil, err
			}
			ts = append(ts, t)
		}
		return ts, nil
	}
}

func s2ss(s string) ([]string, error) {
	if s == "" {
		return []string{}, nil
	}
	return csv.NewReader(strings.NewReader(s)).Read()
}

func (opts *options) declareFlags(fs *flag.FlagSet, parentID string) {
	for i := 0; i < opts.t.NumField(); i++ {
		ptr := opts.v.Field(i).Addr().UnsafePointer()
		sf := opts.t.Field(i)
		usage := docs[parentID+"."+sf.Name]
		switch sf.Type.Kind() {
		case reflect.Bool:
			flagVar(flagVarOpts[bool]{fs.BoolVarP, ptr, sf, strconv.ParseBool, usage})
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			flagVar(flagVarOpts[int64]{fs.Int64VarP, ptr, sf, s2i64, usage})
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			flagVar(flagVarOpts[uint64]{fs.Uint64VarP, ptr, sf, s2u64, usage})
		case reflect.Float32, reflect.Float64:
			flagVar(flagVarOpts[float64]{fs.Float64VarP, ptr, sf, s2f64, usage})
		case reflect.String:
			flagVar(flagVarOpts[string]{fs.StringVarP, ptr, sf, s2s, usage})
		case reflect.Slice:
			switch sf.Type.Elem().Kind() {
			case reflect.Bool:
				flagVar(flagVarOpts[[]bool]{fs.BoolSliceVarP, ptr, sf, toSlice(strconv.ParseBool), usage})
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				flagVar(flagVarOpts[[]int64]{fs.Int64SliceVarP, ptr, sf, toSlice(s2i64), usage})
			case reflect.Float32, reflect.Float64:
				flagVar(flagVarOpts[[]float64]{fs.Float64SliceVarP, ptr, sf, toSlice(s2f64), usage})
			case reflect.String:
				flagVar(flagVarOpts[[]string]{fs.StringSliceVarP, ptr, sf, toSlice(s2s), usage})
			default:
				fmt.Fprintf(os.Stderr, "got: '%#v'; want: []bool|[]int|[]float|[]string fields\n", sf)
				os.Exit(1)
			}
		case reflect.Struct:
			if opts.parentIndex != -1 {
				fmt.Fprintf(os.Stderr, "got: '%#v'; want: exactly one struct field\n", sf)
				os.Exit(1)
			}
			opts.parentIndex = i
		default:
			fmt.Fprintf(os.Stderr, "got: '%#v'; want: bool|int|uint|float|string|slice fields\n", sf)
			os.Exit(1)
		}
	}
}
