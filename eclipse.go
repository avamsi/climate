package eclipse

import (
	"fmt"
	"math/bits"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
)

func overreact(err error) {
	if err != nil {
		panic(err)
	}
}

var (
	anyUpperLower = regexp.MustCompile("(.)([A-Z][a-z])")
	lowerUpper    = regexp.MustCompile("([a-z])([A-Z])")
)

func toSnakeCase(s string) string {
	s = anyUpperLower.ReplaceAllString(s, "${1}_${2}")
	s = lowerUpper.ReplaceAllString(s, "${1}_${2}")
	return strings.ToLower(s)
}

func defineFlags(fs *flag.FlagSet, st reflect.Type) {
	for i := 0; i < st.NumField(); i++ {
		sf := st.Field(i)
		sdefault, ok := sf.Tag.Lookup("default")
		usage := sf.Tag.Get("usage")
		switch sf.Type.Kind() {
		case reflect.Bool:
			b := false
			if ok {
				var err error
				b, err = strconv.ParseBool(sdefault)
				overreact(err)
			}
			fs.Bool(toSnakeCase(sf.Name), b, usage)
		case reflect.String:
			if !ok {
				sdefault = ""
			}
			fs.String(toSnakeCase(sf.Name), sdefault, usage)
		default:
			panic(fmt.Sprintf("want: bool or string; got: %v", sf))
		}
	}
}

func copyFlagsToStruct(fs *flag.FlagSet, st reflect.Type, sv reflect.Value) {
	for i := 0; i < st.NumField(); i++ {
		sf := st.Field(i)
		switch sf.Type.Kind() {
		case reflect.Bool:
			sfv, err := fs.GetBool(toSnakeCase(sf.Name))
			overreact(err)
			sv.Field(i).SetBool(sfv)
		case reflect.String:
			sfv, err := fs.GetString(toSnakeCase(sf.Name))
			overreact(err)
			sv.Field(i).SetString(sfv)
		default:
			panic(fmt.Sprintf("want: bool or string; got: %v", sf))
		}
	}
}

func flagsAsStruct(fs *flag.FlagSet, st reflect.Type) reflect.Value {
	sv := reflect.New(st).Elem()
	copyFlagsToStruct(fs, st, sv)
	return sv
}

type input uint8

const (
	inputUnknown input = 1 << iota
	inputParentFlags
	inputArgs
)

func copyCallableToCmd(in input, ct reflect.Type, cv reflect.Value, cmd *cobra.Command) {
	expNumIn := bits.OnesCount8(uint8(in)) + 1
	if ct.NumIn() != expNumIn || ct.IsVariadic() || ct.NumOut() != 0 {
		panic(fmt.Sprintf("want: callable with exactly %v non-variadic inputs and no outputs; got: %v", expNumIn, ct))
	}
	offset := 0
	if in&inputParentFlags != 0 {
		offset += 1
	}
	flags := ct.In(offset)
	if flags.Kind() != reflect.Struct {
		panic(fmt.Sprintf("want: struct; got: %v", flags))
	}
	if in&inputArgs != 0 {
		args := ct.In(offset + 1)
		if !(args.Kind() == reflect.Slice && args.Elem().Kind() == reflect.String) {
			panic(fmt.Sprintf("want: slice of strings; got: %v", args))
		}
	}
	fs := cmd.Flags()
	defineFlags(fs, flags)
	cmd.Run = func(cmd *cobra.Command, args []string) {
		inputs := []reflect.Value{flagsAsStruct(fs, flags)}
		if in&inputArgs != 0 {
			inputs = append(inputs, reflect.ValueOf(args))
		}
		cv.Call(inputs)
	}
}

func structAsCmd(st reflect.Type) *cobra.Command {
	cmd := &cobra.Command{Use: toSnakeCase(st.Name())}
	fs := cmd.PersistentFlags()
	defineFlags(fs, st)
	sv := reflect.New(st).Elem()
	cmd.PersistentPreRun = func(*cobra.Command, []string) {
		copyFlagsToStruct(fs, st, sv)
	}
	for i := 0; i < st.NumMethod(); i++ {
		m, mv := st.Method(i), sv.Method(i)
		if m.Name == "Execute" {
			copyCallableToCmd(inputParentFlags, m.Type, mv, cmd)
		} else {
			subCmd := &cobra.Command{Use: toSnakeCase(m.Name)}
			copyCallableToCmd(inputParentFlags|inputArgs, m.Type, mv, subCmd)
			cmd.AddCommand(subCmd)
		}
	}
	return cmd
}

func Execute(i interface{}) {
	t := reflect.TypeOf(i)
	switch t.Kind() {
	case reflect.Func:
		cmd := &cobra.Command{Use: toSnakeCase(t.Name())}
		copyCallableToCmd(inputArgs, t, reflect.ValueOf(i), cmd)
		overreact(cmd.Execute())
	case reflect.Struct:
		overreact(structAsCmd(t).Execute())
	default:
		panic(fmt.Sprintf("want: func or struct; got: %v", t))
	}
}
