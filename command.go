package climate

import (
	"reflect"

	"github.com/avamsi/climate/internal"

	"github.com/avamsi/ergo"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type command struct {
	delegate cobra.Command
}

func newCommand(name string, md *internal.Metadata) *command {
	delegate := cobra.Command{
		Use:     md.Usage(name),
		Aliases: md.Aliases(),
		Short:   md.Short(),
		Long:    md.Long(),
	}
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
		cmd    = newCommand(fcb.name, fcb.md)
		i      = 0
		n      = fcb.t().NumIn()
		inOpts *reflect.Value
		inArgs = false
	)
	// We support the signatures func([opts *T], [args []string]) [(err error)],
	// which is to say all of opts, args and error are optional. If opts is
	// present, T must be a struct (and we use its fields as flags).
	// TODO: maybe support variadic, array and normal string arguments too.
	if i < n {
		t := fcb.t().In(i)
		if typeIsStructPointer(t) {
			r := reflection{ptr: &reflection{ot: t}}
			i++
			opts := &options{
				r,
				nil, // no parent
				cmd.delegate.Flags(),
				fcb.md.LookupType(t.Elem()),
			}
			opts.declare()
			inOpts = r.ptr.v()
		}
	}
	if i < n && typeIsStringSlice(fcb.t().In(i)) {
		i++
		inArgs = true
	} else {
		cmd.delegate.Args = cobra.ExactArgs(0)
	}
	outErr := fcb.t().NumOut() == 1 && typeIsError(fcb.t().Out(0))
	if i != n || fcb.t().IsVariadic() || (fcb.t().NumOut() != 0 && !outErr) {
		ergo.Panicf("not func([*struct], [[]string]) [error]: %q", fcb.t())
	}
	cmd.delegate.RunE = func(_ *cobra.Command, args []string) error {
		var in []reflect.Value
		if inOpts != nil {
			in = append(in, *inOpts)
		}
		if inArgs {
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
	cmd := newCommand(scb.t().Name(), scb.md)
	opts := &options{
		scb.reflection,
		scb.parent,
		cmd.delegate.PersistentFlags(),
		scb.md,
	}
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
