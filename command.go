package clifr

import (
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/spf13/cobra"
)

func addr[T any](arg T) *T {
	return &arg
}

type command struct {
	opts     *options
	cobraCmd *cobra.Command
}

type registry struct {
	m map[string]*command
}

func (reg *registry) get(t reflect.Type) *command {
	if t.PkgPath() == "" || t.Name() == "" {
		return nil
	}
	return reg.m[t.PkgPath()+"."+t.Name()]
}

func (reg *registry) put(cmd *command) {
	if cmd.opts.t.PkgPath() != "" && cmd.opts.t.Name() != "" {
		reg.m[cmd.opts.t.PkgPath()+"."+cmd.opts.t.Name()] = cmd
	}
}

var reg = registry{m: map[string]*command{}}

func parseMethod(m reflect.Method) (optsT *reflect.Type, argsIn bool) {
	if m.Type.IsVariadic() {
		fmt.Fprintf(os.Stderr, "got: '%#v'; want: method with non-variadic inputs\n", m.Type)
		os.Exit(1)
	}
	i := 1
	if i < m.Type.NumIn() && m.Type.In(i).Kind() == reflect.Struct {
		optsT = addr(m.Type.In(i))
		i++
	}
	if i < m.Type.NumIn() {
		argsT := m.Type.In(i)
		if argsT.Kind() == reflect.Slice && argsT.Elem().Kind() == reflect.String {
			argsIn = true
			i++
			if m.Name == "Execute" {
				fmt.Fprintf(os.Stderr, "got: '%#v'; "+
					"want: Execute method to not have slice of strings input\n", m.Type)
				os.Exit(1)
			}
		}
	}
	if i != m.Type.NumIn() {
		fmt.Fprintf(os.Stderr, "got: '%#v'; want: method with an optional struct input, "+
			"followed by an optional slice of strings input\n", m.Type)
		os.Exit(1)
	}
	return optsT, argsIn
}

func copyMethodToCobraCmd(m reflect.Method, s reflect.Value, cobraCmd *cobra.Command, parentID string) {
	optsT, argsIn := parseMethod(m)
	var opts *options
	if optsT != nil {
		if (*optsT).Name() != "" {
			parentID = (*optsT).PkgPath() + "." + (*optsT).Name()
		}
		opts = newOptions(*optsT, cobraCmd.Flags(), parentID)
	}
	cobraCmd.Run = func(_ *cobra.Command, args []string) {
		inputs := []reflect.Value{s}
		if opts != nil {
			inputs = append(inputs, opts.v)
		}
		if !argsIn && len(args) > 0 {
			fmt.Fprintf(os.Stderr, "got unrecognized inputs: '%#v'; please run --help\n", args)
			os.Exit(1)
		}
		inputs = append(inputs, reflect.ValueOf(args))
		outs := m.Func.Call(inputs[:m.Type.NumIn()])
		if n := len(outs); n > 0 {
			var err error
			if m.Type.Out(n - 1).Implements(reflect.TypeOf((*error)(nil)).Elem()) {
				err, _ = outs[n-1].Interface().(error)
				outs = outs[:n-1]
			}
			for _, out := range outs {
				fmt.Println(out.String())
			}
			if err != nil {
				fmt.Fprintln(os.Stderr, err.Error())
				os.Exit(1)
			}
		}
		os.Exit(0)
	}
}

func typeToCommand(t reflect.Type) *command {
	if t.Kind() != reflect.Struct {
		fmt.Fprintf(os.Stderr, "got: '%#v'; want: struct\n", t)
		os.Exit(1)
	}
	if cmd := reg.get(t); cmd != nil {
		return cmd
	}
	parentID := t.PkgPath() + "." + t.Name()
	d := docs[parentID]
	cobraCmd := &cobra.Command{
		Use:   d.usageOr(strings.ToLower(t.Name())),
		Long:  d.Long,
		Short: d.shortOrAutoGen(),
	}
	opts := newOptions(t, cobraCmd.PersistentFlags(), parentID)
	cmd := &command{opts, cobraCmd}
	reg.put(cmd)
	if opts.parentIndex != -1 {
		parent := typeToCommand(t.Field(opts.parentIndex).Type)
		opts.v.Field(opts.parentIndex).Set(parent.opts.v)
		parent.cobraCmd.AddCommand(cobraCmd)
	}
	for i := 0; i < t.NumMethod(); i++ {
		m := t.Method(i)
		tmpCobraCmd := cobraCmd
		parentID := parentID + "." + m.Name
		if m.Name != "Execute" {
			d := docs[parentID]
			tmpCobraCmd = &cobra.Command{
				Use:   d.usageOr(strings.ToLower(m.Name)),
				Long:  d.Long,
				Short: d.shortOrAutoGen(),
			}
			cobraCmd.AddCommand(tmpCobraCmd)
		}
		copyMethodToCobraCmd(m, opts.v, tmpCobraCmd, parentID)
	}
	return cmd
}
