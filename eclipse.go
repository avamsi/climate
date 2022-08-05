package eclipse

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/avamsi/checks"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
)

var (
	anyUpperLower = regexp.MustCompile("(.)([A-Z][a-z])")
	lowerUpper    = regexp.MustCompile("([a-z])([A-Z])")
)

func toSnakeCase(s string) string {
	s = anyUpperLower.ReplaceAllString(s, "${1}_${2}")
	s = lowerUpper.ReplaceAllString(s, "${1}_${2}")
	return strings.ToLower(s)
}

func declareFlag[T bool | int | string](
	flagsF func(string, T, string) *T,
	sf reflect.StructField,
	strconvF func(string) (T, error),
) {
	defaultTag, ok := sf.Tag.Lookup("default")
	var defaultValue T
	if ok {
		defaultValue = checks.Check1(strconvF(defaultTag))
	}
	flagsF(toSnakeCase(sf.Name), defaultValue, sf.Tag.Get("usage"))
}

func validateStructAndDeclareFlags(st reflect.Type, fs *flag.FlagSet, parentSV *reflect.Value) {
	for i := 0; i < st.NumField(); i++ {
		sf := st.Field(i)
		switch sf.Type.Kind() {
		case reflect.Bool:
			declareFlag(fs.Bool, sf, strconv.ParseBool)
		case reflect.Int:
			declareFlag(fs.Int, sf, strconv.Atoi)
		case reflect.String:
			declareFlag(fs.String, sf, func(s string) (string, error) { return s, nil })
		case reflect.Pointer:
			if sf.Name != "_" || sf.Type.Elem().Kind() != reflect.Struct {
				panic(fmt.Sprintf("want: pointer field interpreted as a subcommand "+
					"to be a struct and be named `_`; got: `%v`", sf))
			}
		case reflect.Struct:
			if sf.Anonymous || sf.Type.Name() != parentSV.Type().Name() {
				panic(fmt.Sprintf("want: struct field interpreted as the parent command "+
					"to be of the type `%v` and be named; got: `%v`", parentSV.Type(), sf.Type))
			}
		default:
			panic(fmt.Sprintf("want: bool|int|string; got: `%v`", sf))
		}
	}
}

func populateStruct(sv reflect.Value, fs *flag.FlagSet, parentSV *reflect.Value) {
	for i := 0; i < sv.NumField(); i++ {
		sf := sv.Type().Field(i)
		switch sf.Type.Kind() {
		case reflect.Bool:
			sv.Field(i).SetBool(checks.Check1(fs.GetBool(toSnakeCase(sf.Name))))
		case reflect.Int:
			sv.Field(i).SetInt(int64(checks.Check1(fs.GetInt(toSnakeCase(sf.Name)))))
		case reflect.String:
			sv.Field(i).SetString(checks.Check1(fs.GetString(toSnakeCase(sf.Name))))
		case reflect.Struct:
			sv.Field(i).Set(*parentSV)
		}
	}
}

type inputType uint8

const (
	inputUnknown inputType = 1 << iota
	inputParentFlags
	inputArgs
)

func copyCallableToCmd(it inputType, ct reflect.Type, cv reflect.Value, cmd *cobra.Command) {
	if ct.IsVariadic() || ct.NumOut() != 0 {
		panic(fmt.Sprintf("want: callable with non-variadic inputs and no outputs; got: `%v`", ct))
	}
	i := 0
	if it&inputParentFlags != 0 {
		i += 1
	}
	flagsIn, argsIn := false, false
	flagsT := reflect.TypeOf(struct{}{})
	if i < ct.NumIn() && ct.In(i).Kind() == reflect.Struct {
		flagsT = ct.In(i)
		flagsIn = true
		i += 1
	}
	if i < ct.NumIn() && it&inputArgs != 0 {
		if argsT := ct.In(i); argsT.Kind() == reflect.Slice &&
			argsT.Elem().Kind() == reflect.String {
			argsIn = true
			i += 1
		}
	}
	if i != ct.NumIn() {
		panic(fmt.Sprintf("want: callable with an optional struct input, "+
			"followed by an optional slice of strings input; got: `%v`", ct))
	}
	fs := cmd.Flags()
	validateStructAndDeclareFlags(flagsT, fs, nil)
	cmd.Run = func(cmd *cobra.Command, args []string) {
		inputs := []reflect.Value{}
		if flagsIn {
			sv := reflect.New(flagsT).Elem()
			populateStruct(sv, fs, nil)
			inputs = append(inputs, sv)
		}
		if argsIn {
			inputs = append(inputs, reflect.ValueOf(args))
		}
		cv.Call(inputs)
	}
}

func structAsCmd(st reflect.Type, parentSV *reflect.Value) *cobra.Command {
	cmd := &cobra.Command{Use: toSnakeCase(st.Name())}
	fs := cmd.PersistentFlags()
	validateStructAndDeclareFlags(st, fs, parentSV)
	sv := reflect.New(st).Elem()
	cmd.PersistentPreRun = func(cmd *cobra.Command, _ []string) {
		if cmd.HasParent() {
			parentCmd := cmd.Parent()
			parentCmd.PersistentPreRun(parentCmd, []string{})
		}
		populateStruct(sv, fs, parentSV)
	}
	for i := 0; i < st.NumField(); i++ {
		sf := st.Field(i)
		if sf.Type.Kind() == reflect.Pointer {
			cmd.AddCommand(structAsCmd(sf.Type.Elem(), &sv))
		}
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
		checks.Check0(cmd.Execute())
	case reflect.Struct:
		checks.Check0(structAsCmd(t, nil).Execute())
	default:
		panic(fmt.Sprintf("want: func or struct; got: `%v`", t))
	}
}
