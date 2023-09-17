module github.com/avamsi/climate/cmd/examples

go 1.20

require (
	github.com/avamsi/climate v0.0.0
	github.com/avamsi/climate/cmd/climate v0.0.0
	github.com/google/go-cmp v0.5.9
)

replace (
	github.com/avamsi/climate v0.0.0 => ../..
	github.com/avamsi/climate/cmd/climate v0.0.0 => ../climate
)

require (
	github.com/avamsi/ergo v0.0.0-20230913052809-930ba920a5a0 // indirect
	github.com/fatih/color v1.15.0 // indirect
	github.com/gosuri/uitable v0.0.4 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.19 // indirect
	github.com/mattn/go-runewidth v0.0.15 // indirect
	github.com/rivo/uniseg v0.4.4 // indirect
	github.com/sanity-io/litter v1.5.5 // indirect
	github.com/spf13/cobra v1.7.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	golang.org/x/exp v0.0.0-20230905200255-921286631fa9 // indirect
	golang.org/x/mod v0.12.0 // indirect
	golang.org/x/sys v0.12.0 // indirect
	golang.org/x/text v0.13.0 // indirect
	golang.org/x/tools v0.13.0 // indirect
)
