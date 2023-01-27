package internal

type Metadata struct {
	Doc                         string   // other than Clifr directives
	Aliases, ShortDoc, UsageDoc string   // Clifr directives
	ParamNames                  []string // of functions / methods
}
