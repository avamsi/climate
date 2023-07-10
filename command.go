package climate

import (
	"reflect"

	"github.com/avamsi/climate/internal"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type command struct {
	delegate cobra.Command
}

func newCommand(name string, md *internal.Metadata, params []internal.ParamType) *command {
	delegate := cobra.Command{
		Use:     md.Usage(name, params),
		Aliases: md.Aliases(),
		Short:   md.Short(),
		Long:    md.Long(),
	}
	delegate.Flags().SortFlags = false
	delegate.PersistentFlags().SortFlags = false
	if md != nil {
		delegate.DisableFlagsInUseLine = true
	}
	return &command{delegate}
}

func (cmd *command) addCommand(sub *command) {
	cmd.delegate.AddCommand(&sub.delegate)
}

func (cmd *command) run() error {
	normalize := func(_ *pflag.FlagSet, name string) pflag.NormalizedName {
		return pflag.NormalizedName(internal.NormalizeToKebabCase(name))
	}
	// While we prefer kebab-case for flags, we do support other well-formed,
	// cases through normalization (but only kebab-case shows up in --help).
	cmd.delegate.SetGlobalNormalizationFunc(normalize)
	return cmd.delegate.Execute()
}

type funcCommandBuilder struct {
	name string
	reflection
	md *internal.Metadata
}

func (fcb *funcCommandBuilder) build() *command {
	var (
		cmd    = newCommand(fcb.name, fcb.md, internal.ParamTypes(fcb.t()))
		i      = 0
		n      = fcb.t().NumIn()
		inOpts *reflect.Value
		inArgs = internal.NoParam
	)
	// We support the signatures func([opts *T], [args []string]) [(err error)],
	// which is to say all of opts, args and error are optional. If opts is
	// present, T must be a struct (and we use its fields as flags).
	// TODO: maybe support variadic, array and normal string arguments too.
	if i < n {
		if t := fcb.t().In(i); typeIsStructPointer(t) {
			var (
				r    = reflection{ptr: &reflection{ot: t}}
				opts = &options{
					r,
					nil, // no parent
					cmd.delegate.Flags(),
					fcb.md.LookupType(t.Elem()),
				}
			)
			opts.declare()
			i++
			inOpts = r.ptr.v()
		}
	}
	if i < n {
		switch t := fcb.t().In(i); t.Kind() {
		case reflect.String:
			i++
			inArgs = internal.RequiredParam
			cmd.delegate.Args = cobra.ExactArgs(1)
		case reflect.Pointer, reflect.Array, reflect.Slice:
			if t.Elem().Kind() != reflect.String {
				break
			}
			i++
			switch t.Kind() {
			case reflect.Pointer:
				inArgs = internal.OptionalParam
				cmd.delegate.Args = cobra.MaximumNArgs(1)
			case reflect.Array:
				inArgs = internal.FixedLengthParam
				cmd.delegate.Args = cobra.ExactArgs(t.Len())
			case reflect.Slice:
				inArgs = internal.ArbitraryLengthParam
			}
		}
	} else {
		cmd.delegate.Args = cobra.ExactArgs(0)
	}
	outErr := fcb.t().NumOut() == 1 && typeIsError(fcb.t().Out(0))
	if i != n || fcb.t().IsVariadic() || (fcb.t().NumOut() != 0 && !outErr) {
		internal.Panicf("not func([*struct], [[]string]) [error]: %q", fcb.t())
	}
	cmd.delegate.RunE = func(_ *cobra.Command, args []string) error {
		var in []reflect.Value
		if inOpts != nil {
			in = append(in, *inOpts)
		}
		switch inArgs {
		case internal.RequiredParam:
			in = append(in, reflect.ValueOf(args[0]))
		case internal.OptionalParam:
			var ptr *string
			if len(args) == 1 {
				ptr = &args[0]
			}
			in = append(in, reflect.ValueOf(ptr))
		case internal.FixedLengthParam:
			arr := reflect.New(fcb.t().In(n - 1)).Elem()
			reflect.Copy(arr, reflect.ValueOf(args))
			in = append(in, arr)
		case internal.ArbitraryLengthParam:
			in = append(in, reflect.ValueOf(args))
		}
		out := fcb.v().Call(in)
		if outErr {
			if err := out[0].Interface(); err != nil {
				return err.(error)
			}
		}
		return nil
	}
	return cmd
}

type structCommandBuilder struct {
	reflection
	parent *reflection
	md     *internal.Metadata
}

func (scb *structCommandBuilder) build() *command {
	var (
		cmd  = newCommand(scb.t().Name(), scb.md, nil)
		opts = &options{
			scb.reflection,
			scb.parent,
			cmd.delegate.PersistentFlags(),
			scb.md,
		}
	)
	opts.declare()
	for i := 0; i < scb.ptr.v().NumMethod(); i++ {
		m := scb.ptr.t().Method(i)
		// We only support pointer receivers, skip value receiver methods.
		if _, ok := scb.t().MethodByName(m.Name); ok {
			continue
		}
		var (
			v   = scb.ptr.v().Method(i)
			fcb = &funcCommandBuilder{
				m.Name,
				reflection{ov: &v},
				scb.md.Child(m.Name),
			}
		)
		// TODO: maybe provide an option to default to a subcommand.
		cmd.addCommand(fcb.build())
	}
	return cmd
}
