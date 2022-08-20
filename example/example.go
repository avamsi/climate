package main

import "fmt"

//go:generate eclipse --cmd=Bazel
type Bazel struct {
	// (always|never|sometimes)
	Strip string `default:"always"`
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

//go:generate eclipse --cmd=Mobile
type Mobile struct {
	Bzl Bazel
}

func (m Mobile) Install() {
	fmt.Println("Mobile.Install", m)
}

func (m Mobile) Uninstall() {
	fmt.Println("Mobile.Uninstall", m)
}
