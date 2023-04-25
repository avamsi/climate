module github.com/avamsi/climate/cmd/climate

go 1.20

require (
	github.com/avamsi/climate v0.0.0
	github.com/avamsi/ergo v0.0.0-20230504113207-f481386973ee
	github.com/sanity-io/litter v1.5.5
	golang.org/x/tools v0.8.0
)

replace github.com/avamsi/climate v0.0.0 => ../..

require (
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/spf13/cobra v1.7.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	golang.org/x/mod v0.10.0 // indirect
	golang.org/x/sys v0.7.0 // indirect
	golang.org/x/text v0.9.0 // indirect
)
