package main

import (
	"fmt"

	"github.com/avamsi/eclipse"
)

func Ls(flags struct {
	All  bool
	List bool
}, args []string) {
	fmt.Println("Ls", flags, args)
}

func main() {
	eclipse.Execute(Ls)
}
