package climate

import (
	"context"
	"fmt"
	"reflect"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/avamsi/ergo/assert"
	"github.com/gosuri/uitable"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"golang.org/x/mod/module"

	"github.com/avamsi/climate/internal"
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

func version() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return ""
	}
	if info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	var (
		rev      string
		t        time.Time
		modified bool
	)
	for _, kv := range info.Settings {
		switch kv.Key {
		case "vcs.revision":
			rev = kv.Value[:12]
		case "vcs.time":
			t = assert.Ok(time.Parse(time.RFC3339Nano, kv.Value))
		case "vcs.modified":
			modified = assert.Ok(strconv.ParseBool(kv.Value))
		}
	}
	if t.IsZero() || rev == "" {
		return ""
	}
	if modified {
		rev += "*"
	}
	return module.PseudoVersion("", "", t, rev)
}

func flagUsages(fset *pflag.FlagSet) string {
	t := uitable.New()
	t.Separator = ""
	t.MaxColWidth = 80
	t.Wrap = true
	// TODO: is this leaving extra spaces at the end of certain rows?
	// See cmd/examples/{greet,jj}/main_test.go for some examples.
	fset.VisitAll(func(f *pflag.Flag) {
		var short string
		if f.Shorthand != "" {
			short = fmt.Sprintf("-%v, ", f.Shorthand)
		}
		var (
			qtype, usage = pflag.UnquoteUsage(f)
			value        string
		)
		if qtype != "" {
			qtype += " "
		}
		if _, ok := f.Annotations[nonZeroDefault]; ok {
			value = fmt.Sprintf("(default %v) ", f.DefValue)
		}
		t.AddRow("  ", short, "--", f.Name, " ", qtype, value, " ", usage)
	})
	return t.String()
}

func (cmd *command) run(ctx context.Context) error {
	normalize := func(_ *pflag.FlagSet, name string) pflag.NormalizedName {
		return pflag.NormalizedName(internal.NormalizeToKebabCase(name))
	}
	// While we prefer kebab-case for flags, we do support other well-formed,
	// cases through normalization (but only kebab-case shows up in --help).
	cmd.delegate.SetGlobalNormalizationFunc(normalize)
	cmd.delegate.AddCommand(&cobra.Command{
		Use: "version",
		Run: func(*cobra.Command, []string) {
			fmt.Println(version())
		},
		Hidden: true,
	})
	cmd.delegate.Version = version()
	// Align the flag usages as a table (pflag's FlagUsages already does this to
	// some extent but doesn't align types and default values).
	cobra.AddTemplateFunc("flagUsages", flagUsages)
	t := cmd.delegate.UsageTemplate()
	t = strings.ReplaceAll(t, ".FlagUsages", " | flagUsages")
	cmd.delegate.SetUsageTemplate(t)
	return cmd.delegate.ExecuteContext(ctx)
}

type funcCommandBuilder struct {
	name string
	reflection
	md *internal.Metadata
}

type runSignature struct {
	numIn  int
	inCtx  bool
	inOpts *reflect.Value
	inArgs internal.ParamType
	outErr bool
}

func (fcb *funcCommandBuilder) run(sig *runSignature) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		var in []reflect.Value
		if sig.inCtx {
			in = append(in, reflect.ValueOf(cmd.Context()))
		}
		if sig.inOpts != nil {
			in = append(in, *sig.inOpts)
		}
		switch sig.inArgs {
		case internal.RequiredParam:
			in = append(in, reflect.ValueOf(args[0]))
		case internal.OptionalParam:
			var ptr *string
			if len(args) == 1 {
				ptr = &args[0]
			}
			in = append(in, reflect.ValueOf(ptr))
		case internal.FixedLengthParam:
			arr := reflect.New(fcb.t().In(sig.numIn - 1)).Elem()
			reflect.Copy(arr, reflect.ValueOf(args))
			in = append(in, arr)
		case internal.ArbitraryLengthParam:
			in = append(in, reflect.ValueOf(args))
		}
		out := fcb.v().Call(in)
		if sig.outErr {
			err := out[0].Interface()
			if err == nil { // if _no_ error
				return nil
			}
			if err, ok := err.(*usageError); ok {
				// Let Cobra print both the error and usage information.
				return err
			}
			// err is not a usage error (anymore), so set SilenceUsage to true
			// to prevent Cobra from printing usage information.
			cmd.SilenceUsage = true
			// exitError may just be used to exit with a particular exit code
			// and not necessarily have anything to print.
			if err, ok := err.(*exitError); ok && len(err.errs) == 0 {
				cmd.SilenceErrors = true
			}
			return err.(error)
		}
		return nil
	}
}

func (fcb *funcCommandBuilder) build() *command {
	var (
		cmd    = newCommand(fcb.name, fcb.md, internal.ParamTypes(fcb.t()))
		i      = 0
		n      = fcb.t().NumIn()
		inCtx  bool
		inOpts *reflect.Value
		inArgs = internal.NoParam
	)
	// We support the signatures (excuse the partial [optional] notation)
	// func([ctx context.Context], [opts *T], [args []string]) [(err error)],
	// which is to say all of ctx, opts, args and error are optional. If opts is
	// present, T must be a struct (and we use its fields as flags).
	// TODO: maybe support variadic, array and normal string arguments too.
	if i < n && typeIsContext(fcb.t().In(i)) {
		i++
		inCtx = true
	}
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
		internal.Panicf("not func([context.Context], [*struct], [[]string]) [error]: %v", fcb.t())
	}
	cmd.delegate.RunE = fcb.run(&runSignature{n, inCtx, inOpts, inArgs, outErr})
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
		var (
			m   = scb.ptr.t().Method(i)
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
