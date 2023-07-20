package climate

import (
	"context"
	"runtime"
	"strings"

	"github.com/avamsi/climate/internal"
)

type funcPlan struct {
	reflection
}

func (fp *funcPlan) execute(ctx context.Context, md *internal.Metadata) error {
	var (
		name = runtime.FuncForPC(fp.v().Pointer()).Name()
		dot  = strings.LastIndex(name, ".")
	)
	pkgPath, name := name[:dot], name[dot+1:]
	fcb := &funcCommandBuilder{
		name,
		fp.reflection,
		md.Lookup(pkgPath, name),
	}
	cmd := fcb.build()
	return cmd.run(ctx)
}

type structPlan struct {
	reflection
	subcommands []*structPlan
}

func (sp *structPlan) buildRecursive(parent *reflection, md *internal.Metadata) *command {
	scb := &structCommandBuilder{
		sp.reflection,
		parent,
		md.LookupType(sp.t()),
	}
	cmd := scb.build()
	for _, sub := range sp.subcommands {
		cmd.addCommand(sub.buildRecursive(&sp.reflection, md))
	}
	return cmd
}

func (sp *structPlan) execute(ctx context.Context, m *internal.Metadata) error {
	root := sp.buildRecursive(nil, m) // no parent
	return root.run(ctx)
}
