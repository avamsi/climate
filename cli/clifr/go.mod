module github.com/avamsi/clifr/cli/clifr

go 1.18

require (
	github.com/avamsi/clifr v0.3.0
	github.com/avamsi/ergo v0.2.0
	github.com/sanity-io/litter v1.5.5
	golang.org/x/exp v0.0.0-20230118134722-a68e582fa157
)

replace github.com/avamsi/clifr v0.3.0 => ../..

require (
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/spf13/cobra v1.6.1 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
)
