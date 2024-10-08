package climate

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"runtime/debug"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/avamsi/ergo"
	"github.com/avamsi/ergo/assert"
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
	var (
		b strings.Builder
		t = tabwriter.NewWriter(&b, 0, 0, 0, ' ', 0)
	)
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
		fmt.Fprintf(t, "  %v\t--%v\t %v\t%v \t%v\n", short, f.Name, qtype, value, usage)
	})
	t.Flush()
	return b.String()
}

func versionCommand(name, v string) *cobra.Command {
	help := fmt.Sprintf("Display %v's version information", name)
	return &cobra.Command{
		Use:   "version",
		Short: help,
		Long:  help + ".",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, _ []string) {
			cmd.Println(v)
		},
	}
}

func (cmd *command) run(ctx context.Context) error {
	normalize := func(_ *pflag.FlagSet, name string) pflag.NormalizedName {
		return pflag.NormalizedName(internal.NormalizeToKebabCase(name))
	}
	// While we prefer kebab-case for flags, we do support other well-formed,
	// cases through normalization (but only kebab-case shows up in --help).
	cmd.delegate.SetGlobalNormalizationFunc(normalize)
	if v := version(); v != "" {
		// Add the version subcommand only when the root command already has
		// subcommands (similar to how Cobra does it for help / completion).
		if cmd.delegate.HasSubCommands() {
			cmd.delegate.AddCommand(versionCommand(cmd.delegate.Name(), v))
		}
		cmd.delegate.Version = v
	}
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
			if out[0].IsNil() { // if _no_ error
				return nil
			}
			err := out[0].Interface().(error)
			if uerr := new(usageError); errors.As(err, &uerr) {
				// Let Cobra print both the error and usage information.
				return err
			}
			// err is not a usage error (anymore), so set SilenceUsage to true
			// to prevent Cobra from printing usage information.
			cmd.SilenceUsage = true
			// exitError may just be used to exit with a particular exit code
			// and not necessarily have anything to print.
			if eerr := new(exitError); errors.As(err, &eerr) {
				cmd.SilenceErrors = len(eerr.errs) == 0
			}
			return err
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
		ergo.Panicf("not func([context.Context], [*struct], [[]string]) [error]: %v", fcb.t())
	}
	cmd.delegate.RunE = fcb.run(&runSignature{n, inCtx, inOpts, inArgs, outErr})
	return cmd
}

type structCommandBuilder struct {
	reflection
	parent *reflection
	md     *internal.Metadata
}

func validateNoArgs(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true
	err := cobra.NoArgs(cmd, args)
	if err == nil { // if _no_ error
		return cmd.Help()
	}
	var b strings.Builder
	fmt.Fprintf(&b, "%v", err)
	if suggestions := cmd.SuggestionsFor(args[0]); len(suggestions) > 0 {
		b.WriteString("\n\nDid you mean this?\n")
		for _, s := range suggestions {
			fmt.Fprintf(&b, "\t%v\n", s)
		}
	}
	fmt.Fprintf(&b, "\nRun '%v --help' for usage.", cmd.CommandPath())
	return errors.New(b.String())
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
	// This should ideally be as simple as setting cobra.NoArgs, but for
	// whatever reason, Cobra doesn't really honor that for subcommands
	// (see spf13/cobra#706, spf13/cobra#981) -- so, we do it ourselves.
	cmd.delegate.RunE = validateNoArgs
	// We only make this command "runnable" to validate NoArgs, so hack the
	// usage template and pretend it's not really runnable.
	// Note: Cobra subcommands will inherit any custom attributes set on the
	// parent command, so we need to be careful here to only apply the changes
	// to the parent command and not any subcommands.
	defaultHelpFunc := cmd.delegate.HelpFunc()
	cmd.delegate.SetHelpFunc(func(c *cobra.Command, _ []string) {
		if c == &cmd.delegate {
			t := cmd.delegate.UsageTemplate()
			t = strings.ReplaceAll(t, "{{if .Runnable}}", "{{if false}}")
			c.SetUsageTemplate(t)
		}
		defaultHelpFunc(c, nil)
	})
	return cmd
}
