package eclipse

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
)

func check0(err error) {
	if err != nil {
		panic(err)
	}
}

func check1[T any](arg T, err error) T {
	check0(err)
	return arg
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
				b = check1(strconv.ParseBool(sdefault))
			}
			fs.Bool(toSnakeCase(sf.Name), b, usage)
		case reflect.Int:
			i := 0
			if ok {
				i = check1(strconv.Atoi(sdefault))
			}
			fs.Int(toSnakeCase(sf.Name), i, usage)
		case reflect.String:
			if !ok {
				sdefault = ""
			}
			fs.String(toSnakeCase(sf.Name), sdefault, usage)
		default:
			panic(fmt.Sprintf("want: bool|int|string; got: %v", sf))
		}
	}
}

func copyFlagsToStruct(fs *flag.FlagSet, st reflect.Type, sv reflect.Value) {
	for i := 0; i < st.NumField(); i++ {
		sf := st.Field(i)
		switch sf.Type.Kind() {
		case reflect.Bool:
			sv.Field(i).SetBool(check1(fs.GetBool(toSnakeCase(sf.Name))))
		case reflect.Int:
			sv.Field(i).SetInt(int64(check1(fs.GetInt(toSnakeCase(sf.Name)))))
		case reflect.String:
			sv.Field(i).SetString(check1(fs.GetString(toSnakeCase(sf.Name))))
		default:
			panic(fmt.Sprintf("want: bool|int|string; got: %v", sf))
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
	if ct.IsVariadic() || ct.NumOut() != 0 {
		panic(fmt.Sprintf("want: callable with non-variadic inputs and no outputs; got: %v", ct))
	}
	i := 0
	if in&inputParentFlags != 0 {
		i += 1
	}
	flagsIn, argsIn := false, false
	flagsT := reflect.TypeOf(struct{}{})
	if i < ct.NumIn() && ct.In(i).Kind() == reflect.Struct {
		flagsT = ct.In(i)
		flagsIn = true
		i += 1
	}
	if i < ct.NumIn() && in&inputArgs != 0 {
		if argsT := ct.In(i); argsT.Kind() == reflect.Slice &&
			argsT.Elem().Kind() == reflect.String {
			argsIn = true
			i += 1
		}
	}
	if i != ct.NumIn() {
		panic(fmt.Sprintf("want: callable with an optional struct input, "+
			"followed by an optional slice of strings input; got: %v", ct))
	}
	fs := cmd.Flags()
	defineFlags(fs, flagsT)
	cmd.Run = func(cmd *cobra.Command, args []string) {
		inputs := []reflect.Value{}
		if flagsIn {
			inputs = append(inputs, flagsAsStruct(fs, flagsT))
		}
		if argsIn {
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
		check0(cmd.Execute())
	case reflect.Struct:
		check0(structAsCmd(t).Execute())
	default:
		panic(fmt.Sprintf("want: func or struct; got: %v", t))
	}
}
