module github.com/avamsi/climate/cmd/examples

go 1.20

require (
	github.com/avamsi/climate v0.0.0
	github.com/avamsi/climate/cmd/cligen v0.0.0
	github.com/avamsi/climate/testing v0.0.0
	github.com/google/go-cmp v0.6.0
)

replace (
	github.com/avamsi/climate v0.0.0 => ../..
	github.com/avamsi/climate/cmd/cligen v0.0.0 => ../cligen
	github.com/avamsi/climate/testing v0.0.0 => ../../testing
)

require (
	github.com/avamsi/ergo v0.0.0-20231016191401-d5d3a5a058ff // indirect
	github.com/fatih/color v1.15.0 // indirect
	github.com/gosuri/uitable v0.0.4 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-runewidth v0.0.15 // indirect
	github.com/rivo/uniseg v0.4.4 // indirect
	github.com/sanity-io/litter v1.5.5 // indirect
	github.com/spf13/cobra v1.7.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	golang.org/x/exp v0.0.0-20231006140011-7918f672742d // indirect
	golang.org/x/mod v0.13.0 // indirect
	golang.org/x/sys v0.13.0 // indirect
	golang.org/x/text v0.13.0 // indirect
	golang.org/x/tools v0.14.0 // indirect
)
