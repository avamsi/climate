package main

import (
	"fmt"

	"github.com/avamsi/eclipse"
)

type Bazel struct {
	CompilationMode string `default:"fast"`
}

func (bzl Bazel) Execute(flags struct{}) {
	fmt.Println("Bazel", bzl, flags)
}

func (bzl Bazel) Build(flags struct {
	Config string
}, args []string) {
	fmt.Println("Bazel.Build", bzl, flags, args)
}

func main() {
	eclipse.Execute(Bazel{})
}
