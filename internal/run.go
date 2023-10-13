package internal

import "context"

type Plan interface {
	Execute(context.Context, *Metadata) error
}

type RunOptions struct {
	Metadata *[]byte
}
