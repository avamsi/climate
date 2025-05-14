module github.com/avamsi/climate/cmd/examples

go 1.24

require (
	github.com/avamsi/climate v0.0.0
	github.com/avamsi/climate/testing v0.0.0
	github.com/google/go-cmp v0.7.0
)

tool github.com/avamsi/climate/cmd/cligen

require (
	github.com/avamsi/climate/cmd/cligen v0.0.0 // indirect
	github.com/avamsi/ergo v0.0.0-20250210165756-c19aaccc8346 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/sanity-io/litter v1.5.8 // indirect
	github.com/spf13/cobra v1.9.1 // indirect
	github.com/spf13/pflag v1.0.6 // indirect
	golang.org/x/exp v0.0.0-20250218142911-aa4b98e5adaa // indirect
	golang.org/x/mod v0.24.0 // indirect
	golang.org/x/sync v0.14.0 // indirect
	golang.org/x/text v0.25.0 // indirect
	golang.org/x/tools v0.33.0 // indirect
)

replace (
	github.com/avamsi/climate v0.0.0 => ../..
	github.com/avamsi/climate/cmd/cligen v0.0.0 => ../cligen
	github.com/avamsi/climate/testing v0.0.0 => ../../testing
)
