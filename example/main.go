package main

import (
	"fmt"

	_ "embed"

	"github.com/avamsi/eclipse"
)

type Bazel struct {
	// (always|never|sometimes)
	Strip string `default:"always"`
}

func (bzl Bazel) Execute() {
	fmt.Println("Bazel", bzl)
}

func (bzl Bazel) Build(flags struct {
	CPU string
}, args []string) {
	fmt.Println("Bazel.Build", bzl, flags, args)
}

type Mobile struct {
	Bzl Bazel
}

func (m Mobile) Install() {
	fmt.Println("Mobile.Install", m)
}

func (m Mobile) Uninstall() {
	fmt.Println("Mobile.Uninstall", m)
}

//go:generate eclipse docs --out=eclipse.docs
//go:embed eclipse.docs
var docs []byte

func main() {
	eclipse.Execute(docs, Bazel{}, Mobile{})
}
