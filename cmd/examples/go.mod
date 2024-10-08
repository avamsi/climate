module github.com/avamsi/climate/cmd/examples

go 1.23

toolchain go1.23.0

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
	github.com/avamsi/ergo v0.0.0-20240814022523-bd3d9ec1a8ba // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/sanity-io/litter v1.5.5 // indirect
	github.com/spf13/cobra v1.8.1 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	golang.org/x/exp v0.0.0-20240808152545-0cdaa3abc0fa // indirect
	golang.org/x/mod v0.20.0 // indirect
	golang.org/x/sync v0.8.0 // indirect
	golang.org/x/text v0.17.0 // indirect
	golang.org/x/tools v0.24.0 // indirect
)
