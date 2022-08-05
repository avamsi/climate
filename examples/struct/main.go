package main

import (
	"fmt"

	"github.com/avamsi/eclipse"
)

type Bazel struct {
	Strip string `default:"always" usage:"(always|never|sometimes)"`
	_     *Mobile
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

func main() {
	eclipse.Execute(Bazel{})
}
