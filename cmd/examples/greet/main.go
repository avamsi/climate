package main

import (
	"fmt"

	_ "embed"

	"github.com/avamsi/climate"
)

// Exported struct fields are automatically declared as flags --
//	1. Field names are converted to kebab-case and are used as flag names.
//	   That said, users can pass flags in camelCase, PascalCase, snake_case or
//	   SCREAMING_SNAKE_CASE and everything just works.
//	2. Field types are used as flag types (string, bool, int, etc.).
//	3. "short" field tags are used as short flag names (as is).
//	4. "default" field tags are used as default values (of course, with
//	   automatic type conversion from raw string to the actual field type).
//	5. Field docs / comments are used* as flag usage strings (as is).

type greetOptions struct {
	Greeting string `short:"g" default:"Hello"` // greeting to use
	Name     string `short:"n" default:"World"` // name to greet
}

// Func is automatically converted to a command --
//	1. Param names are converted to kebab-case and used* as part of the usage
//	   string ("command [opts] [args]", for example).
//	2. (Optional) First argument if a struct pointer, is used to declare flags.
//	3. (Optional) Next argument if a string slice is used to collect args.
//	4. Doc is used* as long help string (as is).
//	5. Usage directive is used* to explicitly set the usage string.

// Greet someone.
func greet(opts *greetOptions) {
	fmt.Printf("%s, %s!\n", opts.Greeting, opts.Name)
}

// * These only work if you generate and pass along "metadata" like below --

//go:generate go run github.com/avamsi/climate/cmd/climate --out=md.climate
//go:embed md.climate
var md []byte

func main() {
	climate.Run(climate.Func(greet), climate.Metadata(md))
}
